package svg

import "github.com/bep/gr"

func SVG(mods ...gr.Modifier) *gr.Element {
	e := gr.NewElement("svg")
	gr.Modifiers(mods).Modify(e)
	return e
}

func G(mods ...gr.Modifier) *gr.Element {
	e := gr.NewElement("g")
	gr.Modifiers(mods).Modify(e)
	return e
}

func Rect(mods ...gr.Modifier) *gr.Element {
	e := gr.NewElement("rect")
	gr.Modifiers(mods).Modify(e)
	return e
}
