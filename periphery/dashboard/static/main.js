function init() {
    hideSpinner();
    $(document).on( 'keyup', '#filter', function() {
        var data = $('#data .panel');
        console.log($(this).text())
        var re = new RegExp($(this).text(), "i");
        data.show().filter(function() {
            var name = $(this).find(".bpname").text();
            return !re.test(name);
        }).hide();
    });

    loadBPs();
}

function loadBPs() {
    $.getJSON("/api/bps/", function(data) {
        if (data == 0 || data == null || data == undefined) {
            $("<div class='notification panel'>Error while listing business processes...</div>").hide().appendTo("#data").fadeIn("slow");
        } else {
            var count = 0;
            for (var key in data) {
                count++;
                var fadein = (count > 3) ? 600 : 150 * count;

                var bp = { bpid: key, bpname: data[key] };
                var panel = tmpl("bp_tmpl", bp);
                $(panel).hide().appendTo("#data").fadeIn(fadein);
                loadBPEvents(key);
                loadKPIs(key);
            }
        }
    });
}

function loadBPEvents(bpid) {
    $.getJSON("/api/bps/" + bpid, function(data) {
        if (data == 0 || data == null || data == undefined) {
            $("<div class='notification panel'>Could not fetch events for business process '" + bpid + "'...</div>").hide().appendTo("#data").fadeIn("slow");
        } else {
            var percentages = {};
            for (var key in data) {
                frame = data[key];

        		var status = "unknown";
                switch(frame.status) {
    				case 0:
        				status = "ok";
        				break;
    				case 1:
        				status = "nok";
        				break;
    				default:
        				status = "unknown";
				}

                var percent =  frame.duration_percent;

                if (percentages[status] == null || percentages[status] == undefined) {
                    percentages[status] = 0;
                }
				percentages[status] = percentages[status] + percent;

                // make sure data is visible
                //if (percent < 0.1) {
                //    visible_percent = 0.1
                //}
                //
                // https://codepen.io/cbracco/pen/qzukg
                var f = { state: status, percent: percent, start: frame.start };
                var chart = tmpl("chart_frame_tmpl", f);
                $(chart).hide().appendTo("#" + bpid +"_chart").fadeIn("fast");

				if (frame.status == 1) {
					var annotation = (frame.annotation != "") ? frame.annotation : "<i>-</i>";
					var i = { timestamp: formatTimestamp(frame.start), annotation: annotation, duration: frame.duration };
                	var interruption = tmpl("interruption_tmpl", i);
                	$(interruption).hide().appendTo("#" + bpid +"_interruptions").fadeIn("fast");
				}
            }
            if (percentages["ok"] != null || percentages["ok"] != undefined) {
                var out = Number((percentages["ok"]).toFixed(3));
                $("#" + bpid + "_availability").text(out);
            }
        }
    });
}

function loadKPIs(bpid) {
    $.getJSON("/api/bps/" + bpid + "/kpis", function(data) {
        if (data == 0 || data == null || data == undefined) {
            $("<div class='notification panel'>Error while listing KPIs for business processes " + bpid + "...</div>").hide().appendTo("#data").fadeIn("slow");
        } else {
            for (var key in data) {
                var kpi = { bpid: bpid, kpiid: key, kpiname: data[key] };
                var panel = tmpl("kpi_tmpl", kpi);
                $(panel).hide().appendTo("#" + bpid + "_kpis").fadeIn("fast");
                loadKPIEvents(bpid, key);
            }
        }
    });
}

function loadKPIEvents(bpid, kpiid) {
    $.getJSON("/api/bps/" + bpid + "/kpis/" + kpiid, function(data) {
        if (data == 0 || data == null || data == undefined) {
            $("<div class='notification panel'>Could not fetch events for KPI "+ kpiid + " of business process " + bpid + "'...</div>").hide().appendTo("#data").fadeIn("slow");
        } else {
            var percentages = {};
            for (var key in data) {
                frame = data[key];

        		var status = "unknown";
                switch(frame.status) {
    				case 0:
        				status = "ok";
        				break;
    				case 1:
        				status = "nok";
        				break;
    				default:
        				status = "unknown";
				}

                if (percentages[status] == null || percentages[status] == undefined) {
                    percentages[status] = 0;
                }
				percentages[status] = percentages[status] + frame.duration_percent;

                var f = { bpid: bpid, state: status, percent: frame.duration_percent, start: frame.start };
                var chart = tmpl("chart_frame_tmpl", f);
                $(chart).hide().appendTo("#" + bpid + "_" + kpiid + "_chart").fadeIn("fast");
            }
            if (percentages["ok"] != null || percentages["ok"] != undefined) {
                var out = Number((percentages["ok"]).toFixed(3));
                $("#" + bpid + "_" + kpiid + "_availability").text("~" + out);
            }
        }
    });
}


function showDetails(bpid) {
    var id = "#"+bpid+"_details";
    $(id).parent().find(".bp").find(".bitch").find(".show-details").fadeOut(200);
    $(id).slideDown();
}

function hideDetails(bpid) {
    var id = "#"+bpid+"_details";
    $(id).parent().find(".bp").find(".bitch").find(".show-details").fadeIn(200);
    $(id).slideUp();
}
