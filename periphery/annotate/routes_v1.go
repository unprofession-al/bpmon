package annotate

import wh "github.com/unprofession-al/bpmon/periphery/webhelpers"

func init() {
	routes["v1"] = wh.Leafs{
		"{kind}": wh.Leaf{
			L: wh.Leafs{
				"events": wh.Leaf{
					E: wh.Endpoints{
						"GET": wh.Endpoint{N: "ListEvents", H: ListEvents},
					},
					L: wh.Leafs{
						"{event}": wh.Leaf{
							E: wh.Endpoints{
								"POST": wh.Endpoint{N: "AnnotateEvent", H: AnnotateEvent},
							},
						},
					},
				},
			},
		},
	}
}
