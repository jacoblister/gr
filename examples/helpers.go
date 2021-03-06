package examples

import (
	"fmt"
	"strings"

	"github.com/bep/gr"
	"github.com/bep/gr/attr"
	"github.com/bep/gr/el"
	"github.com/bep/gr/evt"
)

// Example is a wrapper for the examples
func Example(title string, body ...gr.Modifier) *gr.Element {
	mods := gr.Modifiers(body)
	elem := el.Div(gr.CSS("panel", "panel-primary"),
		el.Div(gr.CSS("panel-body"), mods),
		el.Div(gr.CSS("panel-footer"),
			el.Div(
				el.Emphasis(gr.Text("Facebook React in Go: ")),
				el.Anchor(attr.HRef("https://github.com/bep/gr/"),
					gr.Text("https://github.com/bep/gr/")),
			)),
	)

	return elem
}

// Panel creates a panel.
func Panel(title string, body ...gr.Modifier) *gr.Element {
	mods := gr.Modifiers(body)
	return el.Div(gr.CSS("panel", "panel-primary"),
		el.Div(gr.CSS("panel-heading"), el.Header2(gr.Text(title))),
		el.Div(gr.CSS("panel-body"),
			mods),
		el.Div(gr.CSS("panel-footer"),
			el.Div(
				el.Emphasis(gr.Text("Facebook React in Go: ")),
				el.Anchor(attr.HRef("https://github.com/bep/gr/"),
					gr.Text("https://github.com/bep/gr/")),
				el.Anchor(attr.HRef("../basic/"),
					gr.Text(" - More examples")),
			)))

}

func exampleListItem(title, href, text string) gr.Modifier {
	var (
		itemStatus gr.Modifier = gr.Discard
		loc                    = gr.Location()
	)

	if !strings.HasSuffix(href, "/") {
		href += "/"
	}

	if strings.HasSuffix(loc.Path, href) {
		itemStatus = gr.CSS("active")
	}

	href = "../" + href

	return el.Anchor(gr.CSS("list-group-item"), itemStatus, attr.HRef(href), gr.Text(text))

}

// Alert creates a Bootstrap alert element.
func Alert(classifier string, body gr.Modifier) *gr.Element {
	e := el.Div(
		gr.CSS("alert", "alert-"+classifier),
		//el.Anchor(attr.HRef("#"),
		//	gr.CSS("close"), gr.Data("dismiss", "alert"), gr.Aria("label", "close"),
		//	gr.Text("Close")),
		body)
	return e
}

// ClickCounter is a reusable components to use in composition examples.
// This is just copy-paste from the click counter example. Consider making something else.
type ClickCounter struct {
	*gr.This
}

// GetInitialState implements the StateInitializer interface.
func (c ClickCounter) GetInitialState() gr.State {
	return gr.State{"counter": 0}
}

// Render implements the Renderer interface.
func (c ClickCounter) Render() gr.Component {
	counter := c.State()["counter"]
	message := fmt.Sprintf(" Click me! Number of clicks: %v", counter)

	return el.Div(
		el.Button(
			gr.CSS("btn", "btn-lg", "btn-primary"),
			gr.Text(message),
			evt.Click(c.onClick)))
}

func (c ClickCounter) onClick(event *gr.Event) {
	c.SetState(gr.State{"counter": c.State().Int("counter") + 1})
}

// ShouldComponentUpdate implements the ShouldComponentUpdate interface.
func (c ClickCounter) ShouldComponentUpdate(
	nextProps gr.Props, nextState gr.State) bool {

	return c.State().HasChanged(nextState, "counter")
}

// ComponentDidMount implements the ComponentDidMount interface.
func (c ClickCounter) ComponentDidMount() {
	println("ClickCounter: ComponentDidMount")
}
