package dashboard

import wh "github.com/unprofession-al/bpmon/periphery/webhelpers"

func init() {
	routes["v1"] = wh.Leafs{
		"bps": wh.Leaf{
			E: wh.Endpoints{
				"GET": wh.Endpoint{N: "ListBPs", H: ListBPsHandler},
			},
			L: wh.Leafs{
				"{bp}": wh.Leaf{
					E: wh.Endpoints{
						"GET": wh.Endpoint{N: "GetBPEvents", H: GetBPTimelineHandler},
					},
					L: wh.Leafs{
						"kpis": wh.Leaf{
							E: wh.Endpoints{
								"GET": wh.Endpoint{N: "ListKPIs", H: ListKPIsHandler},
							},
							L: wh.Leafs{
								"{kpi}": wh.Leaf{
									E: wh.Endpoints{
										"GET": wh.Endpoint{N: "GetKPIEvents", H: GetKPITimelineHandler},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
