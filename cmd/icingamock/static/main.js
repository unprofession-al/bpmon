function loadEnv(name) {
    $.getJSON("/api/envs/" + name, function(data) {
        var size = Object.keys(data).length;
        if (size == 0 || size == undefined) {
            $("<div>No environment found...</div>").hide().appendTo("#data").fadeIn(fadein);
        } else {
            var count = 0;
            for (var key in data) {
                count++;
                var fadein = (count > 3) ? 600 : 150 * count;
                var host = { hostname: key, s: data[key] };
                var pannel = tmpl("host_tmpl", host);
                $(pannel).hide().appendTo("#data").fadeIn(fadein);
            }
        }
    });
}


function toggleState(env, host, service) {
    var state = $("#" + host + "_" + service).find(".state").text();
    var newstate = (state == 0) ? 1 : 0
            console.log(newstate)
    $.ajax({
        type: "POST",
        url: "/api/envs/" + env + "/hosts/" + host + "/services/" + service + "?state=" + newstate,
        success: function(data){
            var addClass = (newstate == 0) ? "good" : "bad";
            var removeClass = (newstate != 0) ? "good" : "bad";
            $("#" + host + "_" + service).addClass(addClass);
            $("#" + host + "_" + service).removeClass(removeClass);
            $("#" + host + "_" + service).find(".state").text(newstate);
        }
    });
}


function init() {
    loadEnv("_")
}

