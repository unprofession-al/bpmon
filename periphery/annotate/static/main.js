function loadEvents() {
    $.getJSON("/api/v1/bp/events" + window.location.search, function(data) {
        hideSpinner();
        if (data == 0 || data === null || data === undefined) {
            $("<div class='notification panel'>No unannotated events found...</div>").hide().appendTo("#data").fadeIn(fadein);
            $("#count").html("0");
        } else {
            var count = 0;
            for (var key in data) {
                count++;
                var fadein = (count > 3) ? 600 : 150 * count;
                var options = {
                    day: "numeric", year: "numeric", month: "short",
                    day: "numeric", hour: "2-digit", minute: "2-digit", second: "2-digit"
                };
                var date = new Date(data[key].time);
                var dateString =  date.toLocaleTimeString("de-ch", options);

                var event = {id: data[key].id, timestamp: dateString, t: data[key].tags};
                var panel = tmpl("event_tmpl", event);

                $(panel).hide().appendTo("#data").fadeIn(fadein);
            }
            $("#count").html(count);
        }
    });
}

function quickAnnotate(id) {
    var message = $('#quick_annotation').text();
    if (message.lenght == 0 || message == "" || message == undefined) {
        notify("Quick Annotation Text to short...");
    } else {
        $.ajax({
            type: "POST",
            url: "/api/v1/bp/events/" + id,
            data: message,
            success: function(data){
                var url = new URL(window.location.href);
                var days = url.searchParams.get("days");
                loadEvents(days);
                $('#'+id).fadeOut(100);
            }
        });
    }
}

function annotate(id) {
    var message = $('#'+id+'_annotation').text();
    if (message.lenght == 0 || message == "" || message == undefined) {
        notify("Annotation Text to short...");
    } else {
        $.ajax({
            type: "POST",
            url: "/api/v1/bp/events/" + id,
            data: message,
            success: function(data){
                var url = new URL(window.location.href);
                var days = url.searchParams.get("days");
                loadEvents(days);
                $('#'+id).fadeOut(100);
            }
        });
    }
}

function showEditor(event) {
    var id = "#"+event+"_annotation_editor";
    $(id).parent().find(".buttons").fadeOut(100);
    $(id).slideDown();
}

function hideEditor(event) {
    var id = "#"+event+"_annotation_editor";
    $(id).parent().find(".buttons").fadeIn(100);
    $(id).slideUp();
}

function hiddenEditor(event) {
    var id = "#"+event+"_annotation_editor";
    $(id).css("display", "none");
}

function reloadData() {
    var url = new URL(window.location.href);
    var days = url.searchParams.get("days");
    var size = url.searchParams.get("size");
    loadEvents(size, days);
}

function init() {
    reloadData();
}
