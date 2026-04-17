package eval

import (
	"testing"

	"github.com/goccy/go-json"
	cdpruntime "github.com/mafredri/cdp/protocol/runtime"
	. "github.com/smartystreets/goconvey/convey"

	jsonf "github.com/MontFerret/ferret/v2/pkg/encoding/json"
	"github.com/MontFerret/ferret/v2/pkg/runtime"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
)

func TestFunction(t *testing.T) {
	Convey("Function", t, func() {
		Convey(".AsAsync", func() {
			Convey("Should set async=true", func() {
				f := F("return 'foo'").AsAsync()
				args := f.eval(EmptyExecutionContextID)

				So(*args.AwaitPromise, ShouldBeTrue)
			})
		})

		Convey(".AsSync", func() {
			Convey("Should set async=false", func() {
				f := F("return 'foo'").AsAsync()
				args := f.eval(EmptyExecutionContextID)

				So(*args.AwaitPromise, ShouldBeTrue)

				args = f.AsSync().eval(EmptyExecutionContextID)

				So(*args.AwaitPromise, ShouldBeFalse)
			})
		})

		Convey(".AsNamed", func() {
			Convey("When without args", func() {
				Convey("Should generate a wrapper with a given function name", func() {
					name := "getFoo"
					exp := "return 'foo'"
					f := F(exp).AsNamed(name)

					So(f.name, ShouldEqual, name)

					call := f.eval(EmptyExecutionContextID)

					expected := "function " + name + "() {\n" + exp + "\n}"

					So(call.FunctionDeclaration, ShouldEqual, expected)
				})
			})

			Convey("When with args", func() {
				Convey("When a declaration is an expression", func() {
					Convey("Should generate a wrapper with a given function name", func() {
						name := "getFoo"
						exp := "return 'foo'"
						f := F(exp).
							AsNamed(name).
							WithArg("bar").
							WithArg(1)

						So(f.name, ShouldEqual, name)

						call := f.eval(EmptyExecutionContextID)

						expected := "function " + name + "(arg1,arg2) {\n" + exp + "\n}"

						So(call.FunctionDeclaration, ShouldEqual, expected)
					})
				})

				Convey("When a declaration is an arrow function", func() {
					Convey("Should generate a wrapper with a given function name", func() {
						name := "getValue"
						exp := "(el) => el.value"
						f := F(exp).
							AsNamed(name).
							WithArgRef("my_element").
							WithArg(1)

						So(f.name, ShouldEqual, name)

						call := f.eval(EmptyExecutionContextID)

						expected := "function " + name + "() {\n" +
							"const $exp = " + exp + ";\n" +
							"return $exp.apply(this, arguments);\n" +
							"}"

						So(call.FunctionDeclaration, ShouldEqual, expected)
					})
				})

				Convey("When a declaration is a plain function", func() {
					Convey("Should generate a wrapper with a given function name", func() {
						name := "getValue"
						exp := "function getElementValue(el) => el.value"
						f := F(exp).
							AsNamed(name).
							WithArgRef("my_element").
							WithArg(1)

						So(f.name, ShouldEqual, name)

						call := f.eval(EmptyExecutionContextID)

						expected := "function " + name + "() {\n" +
							"const $exp = " + exp + ";\n" +
							"return $exp.apply(this, arguments);\n" +
							"}"

						So(call.FunctionDeclaration, ShouldEqual, expected)
					})
				})
			})
		})

		Convey(".AsAnonymous", func() {
			Convey("When without args", func() {
				Convey("Should generate an anonymous wrapper", func() {
					name := ""
					exp := "return 'foo'"
					f := F(exp).AsNamed("getFoo").AsAnonymous()

					So(f.name, ShouldEqual, name)

					call := f.eval(EmptyExecutionContextID)

					expected := "function() {\n" + exp + "\n}"

					So(call.FunctionDeclaration, ShouldEqual, expected)
				})
			})

			Convey("When with args", func() {
				Convey("When a declaration is an expression", func() {
					Convey("Should generate an anonymous wrapper", func() {
						name := ""
						exp := "return 'foo'"
						f := F(exp).
							AsNamed("getFoo").
							AsAnonymous().
							WithArg("bar").
							WithArg(1)

						So(f.name, ShouldEqual, name)

						call := f.eval(EmptyExecutionContextID)

						expected := "function(arg1,arg2) {\n" + exp + "\n}"

						So(call.FunctionDeclaration, ShouldEqual, expected)
					})
				})

				Convey("When a declaration is an arrow function", func() {
					Convey("Should NOT generate a wrapper", func() {
						name := ""
						exp := "(el) => el.value"
						f := F(exp).
							AsNamed("getValue").
							AsAnonymous().
							WithArgRef("my_element").
							WithArg(1)

						So(f.name, ShouldEqual, name)

						call := f.eval(EmptyExecutionContextID)

						So(call.FunctionDeclaration, ShouldEqual, exp)
					})
				})

				Convey("When a declaration is a plain function", func() {
					Convey("Should NOT generate a wrapper", func() {
						name := ""
						exp := "function(el) => el.value"
						f := F(exp).
							AsNamed("getValue").
							AsAnonymous().
							WithArgRef("my_element").
							WithArg(1)

						So(f.name, ShouldEqual, name)

						call := f.eval(EmptyExecutionContextID)

						So(call.FunctionDeclaration, ShouldEqual, exp)
					})
				})
			})
		})

		Convey(".CallOn", func() {
			Convey("It should use a given ownerID over ContextID", func() {
				ownerID := cdpruntime.RemoteObjectID("foo")
				contextID := cdpruntime.ExecutionContextID(42)

				f := F("return 'foo'").CallOn(ownerID)
				call := f.eval(contextID)

				So(call.ExecutionContextID, ShouldBeNil)
				So(call.ObjectID, ShouldNotBeNil)
				So(*call.ObjectID, ShouldEqual, ownerID)
			})

			Convey("It should use a given ContextID when ownerID is empty or nil", func() {
				ownerID := cdpruntime.RemoteObjectID("")
				contextID := cdpruntime.ExecutionContextID(42)

				f := F("return 'foo'").CallOn(ownerID)
				call := f.eval(contextID)

				So(call.ExecutionContextID, ShouldNotBeNil)
				So(call.ObjectID, ShouldBeNil)
				So(*call.ExecutionContextID, ShouldEqual, contextID)
			})
		})

		Convey(".WithArgRef", func() {
			Convey("Should add argument with a given RemoteObjectID", func() {
				f := F("return 'foo'")
				id1 := cdpruntime.RemoteObjectID("foo")
				id2 := cdpruntime.RemoteObjectID("bar")
				id3 := cdpruntime.RemoteObjectID("baz")

				f.WithArgRef(id1).WithArgRef(id2).WithArgRef(id3)

				So(f.Length(), ShouldEqual, 3)

				arg1 := f.args[0]
				arg2 := f.args[1]
				arg3 := f.args[2]

				So(*arg1.ObjectID, ShouldEqual, id1)
				So(arg1.Value, ShouldBeNil)
				So(arg1.UnserializableValue, ShouldBeNil)

				So(*arg2.ObjectID, ShouldEqual, id2)
				So(arg2.Value, ShouldBeNil)
				So(arg2.UnserializableValue, ShouldBeNil)

				So(*arg3.ObjectID, ShouldEqual, id3)
				So(arg3.Value, ShouldBeNil)
				So(arg3.UnserializableValue, ShouldBeNil)
			})
		})

		Convey(".WithArgValue", func() {
			Convey("Should add argument with a given Value", func() {
				f := F("return 'foo'")
				val1 := runtime.NewString("foo")
				val2 := runtime.NewInt(1)
				val3 := runtime.NewBoolean(true)

				f.WithArgValue(val1).WithArgValue(val2).WithArgValue(val3)

				So(f.Length(), ShouldEqual, 3)

				arg1 := f.args[0]
				arg2 := f.args[1]
				arg3 := f.args[2]

				So(arg1.ObjectID, ShouldBeNil)
				So(arg1.Value, ShouldResemble, mustEncodeRuntimeValue(t, val1))
				So(arg1.UnserializableValue, ShouldBeNil)

				So(arg2.ObjectID, ShouldBeNil)
				So(arg2.Value, ShouldResemble, mustEncodeRuntimeValue(t, val2))
				So(arg2.UnserializableValue, ShouldBeNil)

				So(arg3.ObjectID, ShouldBeNil)
				So(arg3.Value, ShouldResemble, mustEncodeRuntimeValue(t, val3))
				So(arg3.UnserializableValue, ShouldBeNil)
			})
		})

		Convey(".WithArg", func() {
			Convey("Should add argument with a given any type", func() {
				f := F("return 'foo'")
				val1 := "foo"
				val2 := 1
				val3 := true

				f.WithArg(val1).WithArg(val2).WithArg(val3)

				So(f.Length(), ShouldEqual, 3)

				arg1 := f.args[0]
				arg2 := f.args[1]
				arg3 := f.args[2]

				So(arg1.ObjectID, ShouldBeNil)
				So(arg1.Value, ShouldResemble, mustMarshalAny(t, val1))
				So(arg1.UnserializableValue, ShouldBeNil)

				So(arg2.ObjectID, ShouldBeNil)
				So(arg2.Value, ShouldResemble, mustMarshalAny(t, val2))
				So(arg2.UnserializableValue, ShouldBeNil)

				So(arg3.ObjectID, ShouldBeNil)
				So(arg3.Value, ShouldResemble, mustMarshalAny(t, val3))
				So(arg3.UnserializableValue, ShouldBeNil)
			})
		})

		Convey(".WithArgSelector", func() {
			Convey("Should add argument with a given QuerySelector", func() {
				f := F("return 'foo'")
				val1 := drivers.NewCSSSelector(".foo-bar")
				val2 := drivers.NewCSSSelector("#submit")
				val3 := drivers.NewXPathSelector("//*[@id='q']")

				f.WithArgSelector(val1).WithArgSelector(val2).WithArgSelector(val3)

				So(f.Length(), ShouldEqual, 3)

				arg1 := f.args[0]
				arg2 := f.args[1]
				arg3 := f.args[2]

				So(arg1.ObjectID, ShouldBeNil)
				So(arg1.Value, ShouldResemble, mustMarshalAny(t, val1.String()))
				So(arg1.UnserializableValue, ShouldBeNil)

				So(arg2.ObjectID, ShouldBeNil)
				So(arg2.Value, ShouldResemble, mustMarshalAny(t, val2.String()))
				So(arg2.UnserializableValue, ShouldBeNil)

				So(arg3.ObjectID, ShouldBeNil)
				So(arg3.Value, ShouldResemble, mustMarshalAny(t, val3.String()))
				So(arg3.UnserializableValue, ShouldBeNil)
			})
		})

		Convey(".Err", func() {
			Convey("Should return nil when no encoding error occurred", func() {
				f := F("return 'foo'").WithArg("ok").WithArgValue(runtime.NewInt(1))

				So(f.Err(), ShouldBeNil)
			})

			Convey("Should capture the first error from WithArg without panicking", func() {
				// chan values cannot be JSON-marshaled.
				f := F("return 'foo'").WithArg(make(chan int))

				So(f.Err(), ShouldNotBeNil)
				// Subsequent builder calls should short-circuit without panicking.
				f.WithArg("ok").WithArgValue(runtime.NewInt(1))
				So(f.Length(), ShouldEqual, 0)
			})
		})

		Convey(".String", func() {
			Convey("It should return a function expression", func() {
				exp := "return 'foo'"
				f := F(exp)

				So(f.String(), ShouldEqual, exp)
			})
		})

		Convey(".returnNothing", func() {
			Convey("It should set return by value to false", func() {
				f := F("return 'foo'").returnNothing()
				call := f.eval(EmptyExecutionContextID)

				So(*call.ReturnByValue, ShouldBeFalse)
			})
		})

		Convey(".returnValue", func() {
			Convey("It should set return by value to true", func() {
				f := F("return 'foo'").returnValue()
				call := f.eval(EmptyExecutionContextID)

				So(*call.ReturnByValue, ShouldBeTrue)
			})
		})

		Convey(".returnRef", func() {
			Convey("It should set return by value to false", func() {
				f := F("return 'foo'").returnValue()
				call := f.eval(EmptyExecutionContextID)

				So(*call.ReturnByValue, ShouldBeTrue)

				f.returnRef()

				call = f.eval(EmptyExecutionContextID)

				So(*call.ReturnByValue, ShouldBeFalse)
			})
		})

		Convey(".compile", func() {
			Convey("When Anonymous", func() {
				Convey("When without args", func() {
					Convey("Should generate an expression", func() {
						name := ""
						exp := "return 'foo'"
						f := F(exp)

						So(f.name, ShouldEqual, name)

						call := f.compile(EmptyExecutionContextID)

						expected := "const args = [];\n" +
							"const " + compiledExpName + " = function() {\n" + exp + "\n};\n" +
							compiledExpName + ".apply(this, args);\n"

						So(call.Expression, ShouldEqual, expected)
					})

					Convey("When a function is given", func() {
						Convey("Should generate an expression", func() {
							name := ""
							exp := "() => return 'foo'"
							f := F(exp)

							So(f.name, ShouldEqual, name)

							call := f.compile(EmptyExecutionContextID)

							expected := "const args = [];\n" +
								"const " + compiledExpName + " = " + exp + ";\n" +
								compiledExpName + ".apply(this, args);\n"

							So(call.Expression, ShouldEqual, expected)
						})
					})
				})

				Convey("When with args", func() {
					Convey("Should generate an expression", func() {
						name := ""
						exp := "return 'foo'"
						f := F(exp).WithArg(1).WithArg("test").WithArg([]int{1, 2})

						So(f.name, ShouldEqual, name)

						call := f.compile(EmptyExecutionContextID)

						expected := "const args = [\n" +
							"1,\n" +
							"\"test\",\n" +
							"[1,2],\n" +
							"];\n" +
							"const " + compiledExpName + " = function(arg1,arg2,arg3) {\n" + exp + "\n};\n" +
							compiledExpName + ".apply(this, args);\n"

						So(call.Expression, ShouldEqual, expected)
					})

					Convey("When a function is given", func() {
						Convey("Should generate an expression", func() {
							name := ""
							exp := "() => return 'foo'"
							f := F(exp).WithArg(1).WithArg("test").WithArg([]int{1, 2})

							So(f.name, ShouldEqual, name)

							call := f.compile(EmptyExecutionContextID)

							expected := "const args = [\n" +
								"1,\n" +
								"\"test\",\n" +
								"[1,2],\n" +
								"];\n" +
								"const " + compiledExpName + " = " + exp + ";\n" +
								compiledExpName + ".apply(this, args);\n"

							So(call.Expression, ShouldEqual, expected)
						})
					})
				})
			})
		})
	})
}

func mustEncodeRuntimeValue(t *testing.T, value runtime.Value) json.RawMessage {
	t.Helper()

	raw, err := jsonf.Default.Encode(value)
	if err != nil {
		t.Fatalf("encode runtime value: %v", err)
	}

	return json.RawMessage(raw)
}

func mustMarshalAny(t *testing.T, value any) json.RawMessage {
	t.Helper()

	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal value: %v", err)
	}

	return json.RawMessage(raw)
}
