package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"
)

type (
	publicationWorkflow struct {
		On struct {
			Push struct {
				Branches []string `yaml:"branches"`
				Tags     []string `yaml:"tags"`
			} `yaml:"push"`
			Dispatch struct {
				Inputs map[string]struct {
					Required bool   `yaml:"required"`
					Type     string `yaml:"type"`
				} `yaml:"inputs"`
			} `yaml:"workflow_dispatch"`
		} `yaml:"on"`
		Jobs map[string]publicationJob `yaml:"jobs"`
	}

	publicationJob struct {
		Needs    string `yaml:"needs"`
		If       string `yaml:"if"`
		Strategy struct {
			FailFast *bool  `yaml:"fail-fast"`
			Matrix   string `yaml:"matrix"`
		} `yaml:"strategy"`
		Concurrency struct {
			Group            string `yaml:"group"`
			CancelInProgress *bool  `yaml:"cancel-in-progress"`
		} `yaml:"concurrency"`
		Steps []publicationStep `yaml:"steps"`
	}

	publicationStep struct {
		Name string            `yaml:"name"`
		Run  string            `yaml:"run"`
		With map[string]string `yaml:"with"`
		Env  map[string]string `yaml:"env"`
	}
)

func TestPublicationWorkflowIsolatesReleases(t *testing.T) {
	workflow := readPublicationWorkflow(t)
	if strings.Join(workflow.On.Push.Branches, ",") != "main" ||
		strings.Join(workflow.On.Push.Tags, ",") != "modules/**/v*" {
		t.Fatal("publication must listen for both main and module tag pushes")
	}

	input := workflow.On.Dispatch.Inputs["release_ref"]
	if !input.Required || input.Type != "string" {
		t.Fatal("manual publication must require a release reference")
	}

	job := workflow.Jobs["publish"]
	if job.Strategy.FailFast == nil || *job.Strategy.FailFast {
		t.Fatal("a failed publication must not cancel other matrix entries")
	}

	if job.Concurrency.Group != "publish-module-${{ matrix.tag }}" ||
		job.Concurrency.CancelInProgress == nil || *job.Concurrency.CancelInProgress {
		t.Fatal("duplicate triggers must serialize by release tag without interrupting an active publication")
	}

	if job.Needs != "discover" || job.If != "needs.discover.outputs.count != '0'" ||
		job.Strategy.Matrix != "${{ fromJSON(needs.discover.outputs.matrix) }}" {
		t.Fatal("publication must consume discovery and skip empty selections")
	}

	tooling := findPublicationStep(t, job, "Checkout publication tooling")

	source := findPublicationStep(t, job, "Checkout released module source")
	if tooling.With["ref"] != "${{ github.workflow_sha }}" ||
		source.With["ref"] != "${{ matrix.commit }}" || source.With["path"] != "source" {
		t.Fatal("backfills must keep workflow tooling separate from the selected release source")
	}
}

func TestPublicationWorkflowReusesBranchAndPullRequest(t *testing.T) {
	workflow := readPublicationWorkflow(t)
	job := workflow.Jobs["publish"]
	push := findPublicationStep(t, job, "Commit and push publication branch")
	pr := findPublicationStep(t, job, "Create or reuse Barn pull request")

	const tag = "modules/db/redis/v1.0.0-rc.5"
	const branch = "publish/contrib/" + tag

	for _, step := range []publicationStep{push, pr} {
		if step.Env["PUBLICATION_BRANCH"] != "publish/contrib/${{ matrix.tag }}" ||
			step.Env["RELEASE_TAG"] != "${{ matrix.tag }}" {
			t.Fatalf("%s must use the matrix tag for publication identity", step.Name)
		}
	}

	root := t.TempDir()
	origin := filepath.Join(root, "origin.git")
	runWorkflowGit(t, root, "init", "--bare", "--initial-branch=main", origin)
	runWorkflowGit(t, root, "clone", origin, "seed")
	seed := filepath.Join(root, "seed")
	runWorkflowGit(t, seed, "config", "user.name", "Publication Test")
	runWorkflowGit(t, seed, "config", "user.email", "publication-test@example.com")
	runWorkflowGit(t, seed, "commit", "--allow-empty", "-m", "initial")
	runWorkflowGit(t, seed, "push", "origin", "main")

	bin := filepath.Join(root, "bin")
	if err := os.Mkdir(bin, 0o755); err != nil {
		t.Fatal(err)
	}

	// Only GitHub is simulated; the workflow's real shell steps push to a local
	// bare repository twice, each time from a fresh checkout of Barn's main.
	gh := `#!/usr/bin/env bash
set -euo pipefail
[[ " $* " == *" --head $PUBLICATION_BRANCH "* ]]
[[ " $* " == *" --repo MontFerret/barn "* ]]
case "$1 $2" in
  'pr list')
    if [[ -f "$FAKE_PR_STATE" ]]; then
      cat "$FAKE_PR_STATE"
    fi
    ;;
  'pr create')
    echo created >> "$FAKE_PR_CREATIONS"
    echo https://example.test/barn/pull/1 > "$FAKE_PR_STATE"
    cat "$FAKE_PR_STATE"
    ;;
  *) exit 1 ;;
esac
`
	if err := os.WriteFile(filepath.Join(bin, "gh"), []byte(gh), 0o755); err != nil {
		t.Fatal(err)
	}

	env := append(os.Environ(),
		"PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"),
		"FAKE_PR_STATE="+filepath.Join(root, "pr"),
		"FAKE_PR_CREATIONS="+filepath.Join(root, "creations"),
		"PUBLICATION_BRANCH="+branch,
		"RELEASE_TAG="+tag,
		"PR_BODY=Publication fixture",
		"GH_TOKEN=fixture",
	)
	const record = "registry/modules/montferret/redis/versions/v1.0.0-rc.5.json"

	for _, attempt := range []string{"first", "duplicate"} {
		work := filepath.Join(root, attempt)
		if err := os.Mkdir(work, 0o755); err != nil {
			t.Fatal(err)
		}

		runWorkflowGit(t, work, "clone", origin, "barn")
		barn := filepath.Join(work, "barn")
		file := filepath.Join(barn, filepath.FromSlash(record))
		if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(file, []byte("release record\n"), 0o644); err != nil {
			t.Fatal(err)
		}

		runWorkflowGit(t, barn, "add", "--", "registry/modules")
		runPublicationStep(t, work, env, push)

		output := runPublicationStep(t, work, env, pr)
		if attempt == "duplicate" && !strings.Contains(output, "Updated existing pull request") {
			t.Fatalf("duplicate publication did not reuse its PR: %s", output)
		}
	}

	creations, err := os.ReadFile(filepath.Join(root, "creations"))
	if err != nil {
		t.Fatal(err)
	}

	if string(creations) != "created\n" {
		t.Fatalf("expected one PR creation, got %q", creations)
	}

	content := runWorkflowGit(t, origin, "show", "refs/heads/"+branch+":"+record)
	if content != "release record\n" {
		t.Fatalf("publication branch content = %q", content)
	}
}

func readPublicationWorkflow(t *testing.T) publicationWorkflow {
	t.Helper()

	content, err := os.ReadFile(filepath.Join("..", "..", ".github", "workflows", "publish-module.yml"))
	if err != nil {
		t.Fatal(err)
	}

	var workflow publicationWorkflow
	if err := yaml.Unmarshal(content, &workflow); err != nil {
		t.Fatal(err)
	}

	return workflow
}

func findPublicationStep(t *testing.T, job publicationJob, name string) publicationStep {
	t.Helper()

	for _, step := range job.Steps {
		if step.Name == name {
			return step
		}
	}

	t.Fatalf("workflow step not found: %s", name)

	return publicationStep{}
}

func runWorkflowGit(t *testing.T, directory string, args ...string) string {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = directory

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, output)
	}

	return string(output)
}

func runPublicationStep(t *testing.T, directory string, env []string, step publicationStep) string {
	t.Helper()

	cmd := exec.Command("bash", "-e", "-o", "pipefail")
	cmd.Dir = directory
	cmd.Env = env
	cmd.Stdin = strings.NewReader(step.Run)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s: %v\n%s", step.Name, err, output)
	}

	return string(output)
}
