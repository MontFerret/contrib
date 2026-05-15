package eval

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/page"
	cdpruntime "github.com/mafredri/cdp/protocol/runtime"

	ferretruntime "github.com/MontFerret/ferret/v2/pkg/runtime"
)

type fakeCDPRuntime struct {
	cdp.Runtime
	props map[cdpruntime.RemoteObjectID]*cdpruntime.GetPropertiesReply
}

func (f *fakeCDPRuntime) GetProperties(
	_ context.Context,
	args *cdpruntime.GetPropertiesArgs,
) (*cdpruntime.GetPropertiesReply, error) {
	props, ok := f.props[args.ObjectID]
	if !ok {
		return nil, fmt.Errorf("unexpected GetProperties call for %q", args.ObjectID)
	}

	return props, nil
}

func TestResolverToListNormalizesMixedRemoteArray(t *testing.T) {
	ctx := context.Background()
	resolver := NewResolver(
		&fakeCDPRuntime{
			props: map[cdpruntime.RemoteObjectID]*cdpruntime.GetPropertiesReply{
				"root": {
					Result: []cdpruntime.PropertyDescriptor{
						remoteProp("0", remoteNode("node-1"), true),
						remoteProp("1", remotePrimitive(StringType, `"hello"`), true),
						remoteProp("2", remotePrimitive(NumberType, `42`), true),
						remoteProp("3", remotePrimitive(BooleanType, `true`), true),
						remoteProp("4", remotePrimitive(ObjectType, `null`), true),
						remoteProp("5", remotePlainObject("obj"), true),
					},
				},
				"obj": {
					Result: []cdpruntime.PropertyDescriptor{
						remoteProp("href", remotePrimitive(StringType, `"https://example.com/path"`), true),
					},
				},
			},
		},
		page.FrameID("frame"),
	)
	resolver.SetLoader(ValueLoaderFn(func(
		_ context.Context,
		_ page.FrameID,
		remoteType RemoteObjectType,
		_ RemoteClassName,
		id cdpruntime.RemoteObjectID,
	) (ferretruntime.Value, error) {
		if remoteType != NodeObjectType {
			t.Fatalf("expected node remote type, got %s", remoteType)
		}

		return ferretruntime.NewString("node:" + string(id)), nil
	}))

	list, err := resolver.ToList(ctx, remoteArray("root"))
	if err != nil {
		t.Fatalf("normalize mixed array: %v", err)
	}

	assertListLength(t, ctx, list, 6)
	assertListValue(t, ctx, list, 0, ferretruntime.NewString("node:node-1"))
	assertListValue(t, ctx, list, 1, ferretruntime.NewString("hello"))
	assertListValue(t, ctx, list, 2, ferretruntime.NewFloat(42))
	assertListValue(t, ctx, list, 3, ferretruntime.True)
	assertListValue(t, ctx, list, 4, ferretruntime.None)

	obj := assertObjectAt(t, ctx, list, 5)
	assertObjectValue(t, ctx, obj, "href", ferretruntime.NewString("https://example.com/path"))
}

func TestResolverToValueNormalizesNestedPlainObjects(t *testing.T) {
	ctx := context.Background()
	resolver := NewResolver(
		&fakeCDPRuntime{
			props: map[cdpruntime.RemoteObjectID]*cdpruntime.GetPropertiesReply{
				"root": {
					Result: []cdpruntime.PropertyDescriptor{
						remoteProp("visible", remotePrimitive(StringType, `"yes"`), true),
						remoteProp("hidden", remotePrimitive(StringType, `"no"`), false),
						{Name: "empty", Enumerable: true},
						remoteProp("items", remoteArray("items"), true),
						remoteProp("nested", remotePlainObject("nested"), true),
					},
				},
				"items": {
					Result: []cdpruntime.PropertyDescriptor{
						remoteProp("0", remotePrimitive(NumberType, `1`), true),
						remoteProp("length", remotePrimitive(NumberType, `1`), false),
					},
				},
				"nested": {
					Result: []cdpruntime.PropertyDescriptor{
						remoteProp("name", remotePrimitive(StringType, `"inner"`), true),
					},
				},
			},
		},
		page.FrameID("frame"),
	)

	out, err := resolver.ToValue(ctx, remotePlainObject("root"))
	if err != nil {
		t.Fatalf("normalize object: %v", err)
	}

	obj, ok := out.(*ferretruntime.Object)
	if !ok {
		t.Fatalf("expected object, got %T", out)
	}

	assertObjectValue(t, ctx, obj, "visible", ferretruntime.NewString("yes"))
	assertObjectValue(t, ctx, obj, "hidden", ferretruntime.None)
	assertObjectValue(t, ctx, obj, "empty", ferretruntime.None)

	itemsValue, err := obj.Get(ctx, ferretruntime.NewString("items"))
	if err != nil {
		t.Fatalf("read items: %v", err)
	}

	items, ok := itemsValue.(*ferretruntime.Array)
	if !ok {
		t.Fatalf("expected items array, got %T", itemsValue)
	}

	assertListLength(t, ctx, items, 1)
	assertListValue(t, ctx, items, 0, ferretruntime.NewFloat(1))

	nestedValue, err := obj.Get(ctx, ferretruntime.NewString("nested"))
	if err != nil {
		t.Fatalf("read nested: %v", err)
	}

	nested, ok := nestedValue.(*ferretruntime.Object)
	if !ok {
		t.Fatalf("expected nested object, got %T", nestedValue)
	}

	assertObjectValue(t, ctx, nested, "name", ferretruntime.NewString("inner"))
}

func TestResolverToListPreservesNoneForUnsupportedRemoteObjects(t *testing.T) {
	ctx := context.Background()
	resolver := NewResolver(
		&fakeCDPRuntime{
			props: map[cdpruntime.RemoteObjectID]*cdpruntime.GetPropertiesReply{
				"root": {
					Result: []cdpruntime.PropertyDescriptor{
						remoteProp("0", remoteObject("style", "", "CSSStyleDeclaration"), true),
					},
				},
			},
		},
		page.FrameID("frame"),
	)

	list, err := resolver.ToList(ctx, remoteArray("root"))
	if err != nil {
		t.Fatalf("normalize unsupported object: %v", err)
	}

	assertListLength(t, ctx, list, 1)
	assertListValue(t, ctx, list, 0, ferretruntime.None)
}

func TestResolverToValueStopsRemoteObjectCycles(t *testing.T) {
	ctx := context.Background()
	resolver := NewResolver(
		&fakeCDPRuntime{
			props: map[cdpruntime.RemoteObjectID]*cdpruntime.GetPropertiesReply{
				"root": {
					Result: []cdpruntime.PropertyDescriptor{
						remoteProp("self", remotePlainObject("root"), true),
					},
				},
			},
		},
		page.FrameID("frame"),
	)

	out, err := resolver.ToValue(ctx, remotePlainObject("root"))
	if err != nil {
		t.Fatalf("normalize cycle: %v", err)
	}

	obj, ok := out.(*ferretruntime.Object)
	if !ok {
		t.Fatalf("expected object, got %T", out)
	}

	assertObjectValue(t, ctx, obj, "self", ferretruntime.None)
}

func remoteProp(name string, value cdpruntime.RemoteObject, enumerable bool) cdpruntime.PropertyDescriptor {
	return cdpruntime.PropertyDescriptor{
		Name:       name,
		Value:      &value,
		Enumerable: enumerable,
	}
}

func remotePrimitive(remoteType RemoteType, value string) cdpruntime.RemoteObject {
	return cdpruntime.RemoteObject{
		Type:  string(remoteType),
		Value: json.RawMessage(value),
	}
}

func remoteArray(id cdpruntime.RemoteObjectID) cdpruntime.RemoteObject {
	return remoteObject(id, ArrayObjectType, "Array")
}

func remoteNode(id cdpruntime.RemoteObjectID) cdpruntime.RemoteObject {
	return remoteObject(id, NodeObjectType, "HTMLDivElement")
}

func remotePlainObject(id cdpruntime.RemoteObjectID) cdpruntime.RemoteObject {
	return remoteObject(id, UnknownObjectType, "Object")
}

func remoteObject(
	id cdpruntime.RemoteObjectID,
	subtype RemoteObjectType,
	className string,
) cdpruntime.RemoteObject {
	ref := cdpruntime.RemoteObject{
		Type:     string(ObjectType),
		ObjectID: &id,
	}

	if subtype != UnknownObjectType {
		rawSubtype := string(subtype)
		ref.Subtype = &rawSubtype
	}

	if className != "" {
		ref.ClassName = &className
	}

	return ref
}

func assertListLength(t *testing.T, ctx context.Context, list ferretruntime.List, expected int) {
	t.Helper()

	length, err := list.Length(ctx)
	if err != nil {
		t.Fatalf("read list length: %v", err)
	}

	if int(length) != expected {
		t.Fatalf("expected list length %d, got %d", expected, length)
	}
}

func assertListValue(
	t *testing.T,
	ctx context.Context,
	list ferretruntime.List,
	idx int,
	expected ferretruntime.Value,
) {
	t.Helper()

	got, err := list.At(ctx, ferretruntime.NewInt(idx))
	if err != nil {
		t.Fatalf("read list[%d]: %v", idx, err)
	}

	if ferretruntime.CompareValues(got, expected) != 0 {
		t.Fatalf("list[%d]: expected %v, got %v", idx, expected, got)
	}
}

func assertObjectAt(
	t *testing.T,
	ctx context.Context,
	list ferretruntime.List,
	idx int,
) *ferretruntime.Object {
	t.Helper()

	got, err := list.At(ctx, ferretruntime.NewInt(idx))
	if err != nil {
		t.Fatalf("read list[%d]: %v", idx, err)
	}

	obj, ok := got.(*ferretruntime.Object)
	if !ok {
		t.Fatalf("expected list[%d] object, got %T", idx, got)
	}

	return obj
}

func assertObjectValue(
	t *testing.T,
	ctx context.Context,
	obj *ferretruntime.Object,
	key string,
	expected ferretruntime.Value,
) {
	t.Helper()

	got, err := obj.Get(ctx, ferretruntime.NewString(key))
	if err != nil {
		t.Fatalf("read object[%q]: %v", key, err)
	}

	if ferretruntime.CompareValues(got, expected) != 0 {
		t.Fatalf("object[%q]: expected %v, got %v", key, expected, got)
	}
}
