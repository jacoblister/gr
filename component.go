/*
Copyright 2016 Bjørn Erik Pedersen <bjorn.erik.pedersen@gmail.com> All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package gr

import (
	"errors"
	"fmt"
	"strings"

	"reflect"

	"github.com/bep/gr/support"
	"github.com/gopherjs/gopherjs/js"
)

var (
	react    = js.Global.Get("React")
	reactDOM = js.Global.Get("ReactDOM")
)

func init() {
	if react == js.Undefined || reactDOM == js.Undefined {
		// Require as a fallback
		var err error
		if react, err = support.Require("react"); err != nil {
			panic(fmt.Sprintf("Cannot find React"))
		}
		if reactDOM, err = support.Require("react-dom"); err != nil {
			panic(fmt.Sprintf("Cannot find ReactDOM"))
		}
	}
}

// A Component represents a React JS component.
//
// http://facebook.github.io/react/docs/glossary.html#react-nodes for a reference.
//
// A Component can be either a constructed element (analogous to a ReactElement)
// or a factory (a ReactClass or a ReactFactory). Factories are identified by their
// implementation of the Factory interface.
type Component interface {
	Node() *js.Object
}

// A Factory is a Component that can construct Elements (analogous to a ReactClass or a ReactFactory).
type Factory interface {
	Component
	CreateElement(props Props, children ...Component) *Element
}

// ReactComponent wraps a Facebook React component.
// This component can either be constructed from a Go implementation (see New) or
// loaded from JavaScript (see FromGlobal and Require).
type ReactComponent struct {
	// The React.createClass response.
	node *js.Object

	// Prototype cached for the cloning purpose.
	elementPrototype *js.Object

	// The minimum interface needed to display something.
	r Renderer

	reactClass *reactClass

	// Options
	exportName      string
	globalName      string
	componentConfig ComponentConfig

	// Needs to be created by createElement as opposed to standalone React factories.
	// TODO(bep) figure a way to extract that info from the JS object.
	needsCreate bool
}

// FromGlobal loads a React component from JavaScript's global object
// ("window" for browsers and "GLOBAL" for Node.js)
func FromGlobal(path ...string) *ReactComponent {

	var component *js.Object

	for _, p := range path {
		if component != nil {
			component = component.Get(p)
		} else {
			component = js.Global.Get(p)
		}
	}

	if component == nil || component == js.Undefined {
		panic(fmt.Sprintf("JS component in path %v not found", path))
	}

	// TODO(bep): No concept of a Renderer implementation here. Do we need it?
	return &ReactComponent{node: component, needsCreate: true}
}

// Require loads a module the Node.js way.
// Note that this requires that the require function is present; if in the browser,
// and not in Node.js, try Browserify.
func Require(path ...string) *ReactComponent {
	m, err := support.Require(path...)
	if err != nil {
		panic(err)
	}
	return &ReactComponent{node: m, needsCreate: true}
}

// ComponentConfig is used to add optional static configuration to a component.
type ComponentConfig struct {
	ContextTypesTemplate Context
}

// Option is used to configure a component.
type Option struct {
	action func(*ReactComponent) error

	// Whether to apply this option on the created React component or not.
	preparePhase bool
}

// WithConfig adds optional static configuration to the component.
func WithConfig(config ComponentConfig) Option {
	// This needs to run before createClass
	return Option{preparePhase: true, action: func(r *ReactComponent) error {
		r.componentConfig = config
		return nil
	}}
}

// Export is an option used to mark that the component should be exported to the
// JavaScript world as a Node.js module export.
func Export(name string) Option {
	return Option{action: func(r *ReactComponent) error {
		if name == "" {
			return errors.New("Must provide export name")
		}
		r.exportName = name
		return nil
	}}
}

// Global is an option used to mark that the component should be exported to the
// JavaScript world as a global with the given name.
func Global(name string) Option {
	return Option{action: func(r *ReactComponent) error {
		if name == "" {
			return errors.New("Must provide global name")
		}
		r.globalName = name
		return nil
	}}
}

// Apply the func to the newly created React component.
func Apply(f func(o *js.Object) *js.Object) Option {
	return Option{action: func(r *ReactComponent) error {
		r.node = f(r.node)
		return nil
	}}
}

// NewSimpleRenderer can be used for quickly putting together components that only
// need to implement Renderer with no need of the owner (this) argument.
func NewSimpleRenderer(c Component) Renderer {
	return simpleRenderer{c}
}

type simpleRenderer struct {
	c Component
}

// Implements the Renderer interface.
func (s simpleRenderer) Render() Component {
	return s.c
}

// NewSimpleComponent can be used for quickly putting together components that only
// need to implement Renderer with no need of the owner (this) argument.
// Especially convenient for testing.
func NewSimpleComponent(c Component, options ...Option) *ReactComponent {
	return New(NewSimpleRenderer(c), options...)
}

type reactClass struct {
	*js.Object

	displayName string `js:"displayName"`

	render            *js.Object `js:"render"`
	getDefaultProps   *js.Object `js:"getDefaultProps"`
	getInitialState   *js.Object `js:"getInitialState"`
	getChildContext   *js.Object `js:"getChildContext"`
	childContextTypes js.M       `js:"childContextTypes"`
	contextTypes      js.M       `js:"contextTypes"`

	shouldComponentUpdate     *js.Object `js:"shouldComponentUpdate"`
	componentWillUpdate       *js.Object `js:"componentWillUpdate"`
	componentDidUpdate        *js.Object `js:"componentDidUpdate"`
	componentWillReceiveProps *js.Object `js:"componentWillReceiveProps"`
	componentWillMount        *js.Object `js:"componentWillMount"`
	componentDidMount         *js.Object `js:"componentDidMount"`
	componentWillUnmount      *js.Object `js:"componentWillUnmount"`
}

type delegateRenderer struct {
	delegate func() Component
}

// Render implements the Renderer interface.
func (d delegateRenderer) Render() Component {
	return d.delegate()
}

// NewRenderer creates a Renderer with the provided func as the implementation.
func NewRenderer(renderFunc func() Component) Renderer {
	return delegateRenderer{renderFunc}
}

// New creates a new Component given a Renderer and optional option(s).
// Note that the Renderer is the minimum interface that needs to be implemented,
// but New will perform interface upgrades for other lifecycle interfaces.
func New(r Renderer, options ...Option) *ReactComponent {
	root := &ReactComponent{r: r, reactClass: &reactClass{Object: js.Global.Get("Object").New()}}

	typ := fmt.Sprintf("%T", r)
	displayName := strings.TrimLeft(typ, "*")
	root.reactClass.displayName = displayName

	// TODO(bep)
	// getDefaultProps propTypes https://github.com/bep/gr/issues/23
	// mixins https://github.com/bep/gr/issues/24
	// statics  https://github.com/bep/gr/issues/25

	ts := extractThisSetter(r)

	// Every component needs to render itself.
	root.reactClass.render = makeRenderFunc(ts, displayName, r.Render)

	// Optional lifecycle implementations below.
	if v, ok := r.(StateInitializer); ok {
		root.reactClass.getInitialState = makeStateFunc(ts, v.GetInitialState)
	} else if ts != nil {
		root.reactClass.getInitialState = js.MakeFunc(func(this *js.Object, arguments []*js.Object) interface{} {
			ts.SetThis(this)
			return nil
		})
	}

	if v, ok := r.(ChildContextProvider); ok {
		root.reactClass.getChildContext, root.reactClass.childContextTypes = makeChildContextFunc(ts, v.GetChildContext)
	}

	if v, ok := r.(ShouldComponentUpdate); ok {
		root.reactClass.shouldComponentUpdate = makeComponentUpdateFunc(ts, v.ShouldComponentUpdate)
	}

	if v, ok := r.(ComponentWillUpdate); ok {
		root.reactClass.componentWillUpdate = makeComponentUpdateVoidFunc(ts, v.ComponentWillUpdate)
	}

	if v, ok := r.(ComponentDidUpdate); ok {
		root.reactClass.componentDidUpdate = makeComponentUpdateVoidFunc(ts, v.ComponentDidUpdate)
	}

	if v, ok := r.(ComponentWillReceiveProps); ok {
		root.reactClass.componentWillReceiveProps = makeComponentPropertyReceiverFunc(ts, v.ComponentWillReceiveProps)
	}

	if v, ok := r.(ComponentWillMount); ok {
		root.reactClass.componentWillMount = makeVoidFunc(ts, v.ComponentWillMount, true)
	}

	if v, ok := r.(ComponentDidMount); ok {
		root.reactClass.componentDidMount = makeVoidFunc(ts, v.ComponentDidMount, true)
	}

	if v, ok := r.(ComponentWillUnmount); ok {
		root.reactClass.componentWillUnmount = makeVoidFunc(ts, v.ComponentWillUnmount, true)
	}

	for _, opt := range options {
		if !opt.preparePhase {
			continue
		}
		err := opt.action(root)
		if err != nil {
			panic(err)
		}
	}

	root.handleOptionsOnPrepare()

	class := react.Call("createClass", root.reactClass.Object)

	root.node = react.Call("createFactory", class)

	for _, opt := range options {
		if opt.preparePhase {
			continue
		}
		err := opt.action(root)
		if err != nil {
			panic(err)
		}
	}

	root.handleOptionsOnCreate()

	return root
}

// CreateIfNeeded evaluates the given Component and returns an Element, creating
// a new instance if needed. This is a convenience method; if you need to pass
// properties, use the factory directly.
func CreateIfNeeded(c Component) *Element {
	switch v := c.(type) {
	case *Element:
		return v
	case Factory:
		return v.CreateElement(nil)
	default:
		return NewPreparedElement(c.Node())
	}
}

// Node implements the Component interface.
func (r *ReactComponent) Node() *js.Object {
	return r.node
}

// Render the Component in the DOM with the given element ID and props.
func (r *ReactComponent) Render(elementID string, props Props) {
	container := js.Global.Get("document").Call("getElementById", elementID)
	elem := r.CreateElement(props)

	// TODO(bep) evaluate if the need the "this" returned on render.
	reactDOM.Call("render", elem.Node(), container)
}

// CreateElement implements the Factory interface.
// TODO(bep) consolidate and clean
func (r *ReactComponent) CreateElement(props Props, children ...Component) *Element {
	return &Element{properties: props, children: children, elFactory: createElementElementFactory(r)}
}

// CloneElement will, provided that an element has already been created for this component, clone that element with the
// original element's props with the new props merged in shallowly.
// New children will replace existing children.
// If this is the first invocation, a new element will be created.
// This may be be slightly faster when creating elements in a tight loop.
//
// See https://facebook.github.io/react/docs/top-level-api.html#react.cloneelement
func (r *ReactComponent) CloneElement(props Props, children ...Component) *Element {
	return &Element{properties: props, children: children, elFactory: cloneOrCreateElementElementFactory(r)}
}

func cloneOrCreateElementElementFactory(r *ReactComponent) func(e *Element) *js.Object {
	return func(e *Element) *js.Object {
		if r.elementPrototype == nil {
			r.elementPrototype = createOrInvoke(r.Node(), e, r.needsCreate)
			return r.elementPrototype
		}
		return cloneElement(r.elementPrototype, e)
	}
}

func createElementElementFactory(r *ReactComponent) func(e *Element) *js.Object {
	return func(e *Element) *js.Object {
		elem := createOrInvoke(r.Node(), e, r.needsCreate)
		r.elementPrototype = elem
		return elem
	}
}

func cloneElement(prototype *js.Object, e *Element) *js.Object {
	var elem *js.Object

	var args []interface{}

	if len(e.children) > 0 {
		for _, c := range e.children {
			args = append(args, c.Node())
		}
	}

	elem = react.Call("cloneElement", prototype, e.properties, args)

	return elem
}

func createOrInvoke(node *js.Object, e *Element, needsCreate bool) *js.Object {
	var elem *js.Object

	var args []interface{}

	if len(e.children) > 0 {
		for _, c := range e.children {
			args = append(args, c.Node())
		}
	}

	if needsCreate {
		elem = react.Call("createElement", node, e.properties, args)
	} else {

		elem = node.Invoke(e.properties, args)
	}
	return elem
}

func extractThisSetter(r Renderer) ThisSetter {
	var thisSetter ThisSetter

	rv := reflect.ValueOf(r)

	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	rt := rv.Type()

	if rt.Kind() == reflect.Struct {
		for i := 0; i < rt.NumField(); i++ {
			fv := rv.Field(i)
			if fv.CanInterface() {
				if init, ok := fv.Interface().(ThisSetter); ok {
					if fv.IsNil() {
						newVal := reflect.New(rt.Field(i).Type.Elem())
						fv.Set(newVal)
						thisSetter = newVal.Interface().(ThisSetter)
					} else {
						thisSetter = init
					}
					// We should maybe check for others and report an error, but for now the first one wins.
					break
				}
			}

		}

	}

	return thisSetter
}

func makeComponentUpdateFunc(ts ThisSetter, f func(c Cops) bool) *js.Object {
	return js.MakeFunc(func(this *js.Object, arguments []*js.Object) interface{} {
		if ts != nil {
			ts.SetThis(this)
		}
		return f(extractComponentUpdateArgs(arguments))
	})
}

func makeComponentUpdateVoidFunc(ts ThisSetter, f func(c Cops)) *js.Object {
	return js.MakeFunc(func(this *js.Object, arguments []*js.Object) interface{} {
		if ts != nil {
			ts.SetThis(this)
		}
		f(extractComponentUpdateArgs(arguments))
		return nil
	})
}

func makeComponentPropertyReceiverFunc(ts ThisSetter, f func(c Cops)) *js.Object {
	return js.MakeFunc(func(this *js.Object, arguments []*js.Object) interface{} {
		if ts != nil {
			ts.SetThis(this)
		}
		data := extractComponentUpdateArgs(arguments)
		f(data)
		return nil
	})
}

func extractComponentUpdateArgs(arguments []*js.Object) Cops {
	var (
		props   Props
		state   State
		context Context
	)

	if len(arguments) > 0 && arguments[0] != nil {
		props = objectToMap(arguments[0])

	}
	if len(arguments) > 1 && arguments[1] != nil {
		state = objectToMap(arguments[1])
	}
	if len(arguments) > 2 && arguments[2] != nil {
		context = objectToMap(arguments[2])
	}

	return Cops{Props: props, State: state, Context: context}
}

func makeVoidFunc(ts ThisSetter, f func(), assumeBlocking bool) *js.Object {
	return js.MakeFunc(func(this *js.Object, arguments []*js.Object) interface{} {
		if ts != nil {
			ts.SetThis(this)
		}
		if assumeBlocking {
			go func() {
				f()
			}()
		} else {
			f()
		}
		return nil
	})
}

func makeStateFunc(ts ThisSetter, f func() State) *js.Object {
	return js.MakeFunc(func(this *js.Object, arguments []*js.Object) interface{} {
		if ts != nil {
			ts.SetThis(this)
		}
		return f()
	})
}

func makeChildContextFunc(ts ThisSetter, f func() Context) (*js.Object, js.M) {

	getChildContext := js.MakeFunc(func(this *js.Object, arguments []*js.Object) interface{} {
		if ts != nil {
			ts.SetThis(this)
		}
		return f()
	})

	childContextTypes := extractPropTypesFromTemplate(f())

	return getChildContext, childContextTypes
}

func extractPropTypesFromTemplate(t map[string]interface{}) js.M {
	propTypes := js.M{}

	for k, v := range t {
		switch v.(type) {
		case string:
			propTypes[k] = react.Get("PropTypes").Get("string")
		case int:
			propTypes[k] = react.Get("PropTypes").Get("number")
		default:
			// See: https://facebook.github.io/react/docs/reusable-components.html
			// TODO(bep): Reconsider all of this.
			panic("Context type not implemented")
		}
	}

	return propTypes
}

type incrementer struct {
	counter int
}

func (i *incrementer) next() int {
	i.counter++
	return i.counter
}

func makeRenderFunc(ts ThisSetter, s string, f func() Component) *js.Object {

	return js.MakeFunc(func(this *js.Object, arguments []*js.Object) interface{} {
		if ts != nil {
			ts.SetThis(this)
		}

		comp := f()

		if comp == nil {
			return nil
		}

		that := NewThis(this)

		// TODO(bep) refactor
		if e, ok := comp.(*Element); ok {
			addEventListeners(ts, comp, that)
			idFactory := &incrementer{}
			addMissingKeys(s, e, idFactory)
		}
		if _, ok := comp.(Factory); ok {
			panic("Render should return a ready-to-use Element.")
		}

		return comp.Node()
	})
}

func makeEventListenerDelegate(ts ThisSetter, el *EventListener, that *This) func(event *js.Object) {
	return func(event *js.Object) {
		if ts != nil {
			ts.SetThis(that.This)
		}
		if el.preventDefault {
			event.Call("preventDefault")
		}
		if el.stopPropagation {
			event.Call("stopPropagation")
		}
		el.listener(&Event{Object: event, This: that})
	}
}

func addEventListeners(ts ThisSetter, c Component, that *This) {
	if e, ok := c.(*Element); ok {
		for _, l := range e.eventListeners {
			l.delegate = makeEventListenerDelegate(ts, l, that)
			e.properties[l.name] = l.delegate

		}
		for _, child := range e.children {
			addEventListeners(ts, child, that)
		}

	}
}

func (r *ReactComponent) handleOptionsOnCreate() {
	if r.exportName != "" {
		exports := js.Module.Get("exports")
		if exports == js.Undefined {
			panic("module.exports not present.")
		}
		exports.Set(r.exportName, r.node)
	}
	if r.globalName != "" {
		js.Global.Set(r.globalName, r.node)
	}

}

func (r *ReactComponent) handleOptionsOnPrepare() {
	if r.componentConfig.ContextTypesTemplate != nil {
		r.reactClass.contextTypes = extractPropTypesFromTemplate(r.componentConfig.ContextTypesTemplate)
	}
}

func addMissingKeys(s string, e *Element, id *incrementer) {

	if !e.dynamic {
		if e.properties == nil {
			e.properties = make(map[string]interface{})
		}
		if _, ok := e.properties["key"]; !ok {
			key := fmt.Sprintf("%s-%d", s, id.next())
			e.properties["key"] = key
		}
	}

	for _, c2 := range e.children {
		if e2, ok := c2.(*Element); ok {
			addMissingKeys(s, e2, id)
		}
	}
}
