var apiBaseURL = "/api/v1/"

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
    $.getJSON(apiBaseURL+"bps/", function(data) {
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
    })
    .fail(function() {
        $("<div class='notification panel'>Error while listing business processes: API could not be reached...</div>").hide().appendTo("#data").fadeIn("slow");
    });
}

function loadBPEvents(bpid) {
    $.getJSON(apiBaseURL+"bps/" + bpid, function(data) {
        if (data == 0 || data == null || data == undefined) {
            $("<div class='notification panel'>Could not fetch events for business process '" + bpid + "'...</div>").hide().appendTo("#data").fadeIn("slow");
        } else {
            var percentages = {};
            var frames = [];
            var events = [];
            for (var key in data) {
                var frame = data[key];
                var status = getStatusString(frame.status);
				var annotation = (frame.annotation != "") ? frame.annotation : "-";
                var percent =  frame.duration_percent;

                if (percentages[status] == null || percentages[status] == undefined) {
                    percentages[status] = 0;
                }
				percentages[status] = percentages[status] + percent;
                var tooltip = "Status: " + status + " &#xa;Start: " + formatTimestamp(frame.start) + " &#xa;End: " + formatTimestamp(frame.end) + "&#xa;&#xa;" + annotation;
                var f = { state: status, displayPercent: percent, percent: percent, start: frame.start, tooltip: tooltip};
                frames.push(f);

				if (frame.status == 1) {
					var e = { timestamp: formatTimestamp(frame.start), annotation: annotation, duration: frame.duration };
                    events.push(e);
				}
            }

            frames = ensureMinimalDisplayPercentage(frames);

            for (var i in frames) {
                var f = frames[i];
                var chart = tmpl("chart_frame_tmpl", f);
                $(chart).hide().appendTo("#" + bpid +"_chart").fadeIn("fast");
            }

            if (events.length > 0) {
                var data = { interruptions: events };
                var interruption = tmpl("interruptions_tmpl", data);
                $(interruption).hide().appendTo("#" + bpid +"_interruptions").fadeIn("fast");
            }

            if (percentages["ok"] != null || percentages["ok"] != undefined) {
                var out = Number((percentages["ok"]).toFixed(3));
                $("#" + bpid + "_availability").text(out);
            }
        }
    })
    .fail(function(jqxhr, textStatus, error) {
        $("<div class='notification panel'>Error while fetching events for business processes " + bpid + ": " + error + "</div>").hide().appendTo("#data").fadeIn("slow");
    });
}

function loadKPIs(bpid) {
    $.getJSON(apiBaseURL+"bps/" + bpid + "/kpis", function(data) {
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
    $.getJSON(apiBaseURL+"bps/" + bpid + "/kpis/" + kpiid, function(data) {
        if (data == 0 || data == null || data == undefined) {
            $("<div class='notification panel'>Could not fetch events for KPI "+ kpiid + " of business process " + bpid + "'...</div>").hide().appendTo("#data").fadeIn("slow");
        } else {
            var percentages = {}
            var frames = [];
            for (var key in data) {
                var frame = data[key];
                var status = getStatusString(frame.status);
                var percent =  frame.duration_percent;

                if (percentages[status] == null || percentages[status] == undefined) {
                    percentages[status] = 0;
                }
				percentages[status] = percentages[status] + percent;

                var tooltip = "Status: " + status + " &#xa;Start: " + formatTimestamp(frame.start) + " &#xa;End: " + formatTimestamp(frame.end);
                var f = { state: status, displayPercent: percent, percent: percent, start: frame.start, tooltip: tooltip};
                frames.push(f);
            }

            frames = ensureMinimalDisplayPercentage(frames);

            for (var i in frames) {
                var f = frames[i];
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

function getStatusString(statusNum) {
    switch(statusNum) {
		case 0:
			return "ok";
		case 1:
			return "nok";
		default:
			return "unknown";
	}
}

function ensureMinimalDisplayPercentage(frames) {
    var minDisplayPercent = 0.2;
    for (var i in frames) {
        var f = frames[i];
        if (f.displayPercent < minDisplayPercent) {
            for (var j in frames) {
                var v = frames[j];
                if (v.displayPercent > (2 * minDisplayPercent)) {
                    frames[j].displayPercent = v.displayPercent - minDisplayPercent;
                    frames[i].displayPercent = f.displayPercent + minDisplayPercent;
                    break;
                }
            }
        }
    }
    return frames;
}

function toggleDetails(bpid) {
    var id = "#"+bpid+"_details";

    if ($(id).is(":visible")) {
        $(id).slideUp();
    } else {
        $(id).slideDown();

    }

    id = "#"+bpid+"_collapser_symbol";
    $(id).toggleClass("fa-angle-up fa-angle-down");
}
