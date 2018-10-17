package dashboard

func init() {
	routes["v1"] = Leafs{
		"annotate": Leaf{
			L: Leafs{
				"{id}": Leaf{
					E: Endpoints{
						"POST": Endpoint{N: "Annotate", H: AnnotateHandler},
					},
				},
			},
		},
		"bps": Leaf{
			E: Endpoints{
				"GET": Endpoint{N: "ListBPs", H: ListBPsHandler},
			},
			L: Leafs{
				"{bp}": Leaf{
					E: Endpoints{
						"GET": Endpoint{N: "GetBPSpans", H: GetBPTimelineHandler},
					},
					L: Leafs{
						"kpis": Leaf{
							E: Endpoints{
								"GET": Endpoint{N: "ListKPIs", H: ListKPIsHandler},
							},
							L: Leafs{
								"{kpi}": Leaf{
									E: Endpoints{
										"GET": Endpoint{N: "GetKPISpans", H: GetKPITimelineHandler},
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
