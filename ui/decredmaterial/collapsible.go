package decredmaterial

import (
	"image/color"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
)

type Collapsible struct {
	isExpanded            bool
	buttonWidget          *widget.Clickable
	expandIcon            *widget.Icon
	headerBackgroundColor color.RGBA
}

func (t *Theme) Collapsible() *Collapsible {
	c := &Collapsible{
		isExpanded:            false,
		headerBackgroundColor: t.Color.Hint,
		expandIcon:            t.navMoreIcon,
		buttonWidget:          new(widget.Clickable),
	}

	return c
}

func (c *Collapsible) layoutHeader(gtx layout.Context, header func(C) D) layout.Dimensions {
	dims := layout.Flex{Spacing: layout.SpaceBetween}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return header(gtx)
		}),
		layout.Rigid(func(gtx C) D {
			return layout.Inset{Right: unit.Dp(20)}.Layout(gtx, func(C) D {
				return c.expandIcon.Layout(gtx, unit.Dp(20))
			})
		}),
	)

	return dims
}

func (c *Collapsible) Layout(gtx layout.Context, header func(C) D, content func(C) D) layout.Dimensions {
	for c.buttonWidget.Clicked() {
		c.isExpanded = !c.isExpanded
	}

	dims := layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return layout.Stack{}.Layout(gtx,
				layout.Stacked(func(gtx C) D {
					gtx.Constraints.Min.X = gtx.Constraints.Max.X
					return c.layoutHeader(gtx, header)
				}),
				layout.Expanded(c.buttonWidget.Layout),
			)
		}),
		layout.Rigid(func(gtx C) D {
			if c.isExpanded {
				return content(gtx)
			}
			return layout.Dimensions{}
		}),
	)
	return dims
}
