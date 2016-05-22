package main

import (
	"time"

	"log"

	"github.com/bep/gr"
	"github.com/bep/gr/el"
	"github.com/bep/gr/examples"
)

// Simple example demonstrating the available lifecycle hooks.
// See also https://facebook.github.io/react/docs/component-specs.html

func main() {

	var (
		lc        gr.Lifecycler = new(lifecycle)
		component               = gr.New(lc)
		props                   = gr.Props{}
		counter                 = 1
	)

	quit := gr.RenderLoop(func() {
		// Only update now and then.
		props["prop"] = counter%100 == 0
		counter++
		component.Render("react", props)
	})

	time.Sleep(10 * time.Second)

	// Stop the render loop
	close(quit)

	if !gr.UnmountComponentAtNode("react") {
		panic("Unmount failed")
	}

	time.Sleep(10 * time.Second)
}

type lifecycle int

// Implements the Renderer interface.
func (l lifecycle) Render(this *gr.This) gr.Component {
	log.Println("Render")
	elem := el.Div(el.Header2(
		gr.Text("Look at the lifecycle events in your console."),
		gr.Style("color", this.State()["color"])))

	return examples.Example("Lifecycle", elem)
}

// Implements the StateInitializer interface.
func (l lifecycle) GetInitialState(this *gr.This) gr.State {
	log.Println("GetInitialState")
	return gr.State{"color": "#ffcc00"}
}

// Implements the ChildContext interface.
func (l lifecycle) GetChildContext(this *gr.This) gr.Context {
	log.Println("GetChildContext")
	return gr.Context{}
}

// Implements the ShouldComponentUpdate interface
func (l lifecycle) ShouldComponentUpdate(this *gr.This, next gr.LifecycleData) bool {
	return this.Props().HasChanged(next.Props, "prop")
}

// Implements the ComponentWillUpdate interface
func (l lifecycle) ComponentWillUpdate(this *gr.This, next gr.LifecycleData) {
	log.Println("ComponentWillUpdate")
}

// Implements the ComponentWillReceiveProps interface
func (l lifecycle) ComponentWillReceiveProps(this *gr.This, data gr.LifecycleData) {
	log.Println("ComponentWillReceiveProps")
}

// Implements the ComponentDidUpdate interface
func (l lifecycle) ComponentDidUpdate(this *gr.This, data gr.LifecycleData) {
	log.Println("ComponentDidUpdate")
}

// Implements the ComponentWillMount interface
func (l lifecycle) ComponentWillMount(this *gr.This) {
	log.Println("ComponentWillMount")
}

// Implements the ComponentDidMount interface
func (l lifecycle) ComponentDidMount(this *gr.This) {
	log.Println("ComponentDidMount")
}

// Implements the ComponentWillUnmount interface
func (l lifecycle) ComponentWillUnmount(this *gr.This) {
	log.Println("ComponentWillUnmount")
}

func init() {
	log.SetFlags(log.Ltime | log.Lmicroseconds)
}
