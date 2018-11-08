package svga

import "github.com/bep/gr"

func Stroke(v interface{}) gr.Modifier {
	return gr.Prop("stroke", v)
}

func Fill(v interface{}) gr.Modifier {
	return gr.Prop("fill", v)
}

func X(v interface{}) gr.Modifier {
	return gr.Prop("x", v)
}

func Y(v interface{}) gr.Modifier {
	return gr.Prop("y", v)
}

func Width(v interface{}) gr.Modifier {
	return gr.Prop("width", v)
}

func Height(v interface{}) gr.Modifier {
	return gr.Prop("height", v)
}
