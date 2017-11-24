var stateNames = [ "ok", "warn", "critical", "unknown" ];

function loadEnv(name) {
    $.getJSON("/api/envs/" + name, function(data) {
        hideSpinner();
        if (data == 0 || data == null || data == undefined) {
            $("<div class='notification panel'>No data in env '" + name + "'...</div>").hide().appendTo("#data").fadeIn(fadein);
        } else {
            var count = 0;
            for (var key in data) {
                var services = data[key]
                for (var s in services) {
                    services[s].color = stateNames[services[s].check_state];
                    services[s].classname = getClassName(key, s);
                }
                count++;
                var fadein = (count > 3) ? 600 : 150 * count;
                var host = { hostname: key, s: services, env: name };
                var panel = tmpl("host_tmpl", host);
                $(panel).hide().appendTo("#data").fadeIn(fadein);
            }
        }
    });
}

function hideSpinner() {
        $("#loading").css("display","none");
}

function toggleState(env, host, service) {
    var classname = getClassName(host, service);
    var before = $("#" + classname).data("check-state");
    console.log(classname);
    var after = ( Number(before) + 1 ) % 4;
    console.log("before " + before + " / after " + after);
    $.ajax({
        type: "POST",
        url: "/api/envs/" + env + "/hosts/" + host + "/services/" + service + "?state=" + after,
        success: function(data){
            var elem = $("#" + classname)
            elem.data("check-state", after);
            var subelem = elem.find(".state");
            subelem.text(stateNames[after]);
            setStateClass(subelem, after);
        }
    });
}

function toggleBool(env, host, service, field) {
    var classname = getClassName(host, service);
    var before = $("#" + classname).data(field);
    var after = ( before != true );
    $.ajax({
        type: "POST",
        url: "/api/envs/" + env + "/hosts/" + host + "/services/" + service + "?" + field + "=" + after,
        success: function(data){
            var elem = $("#" + classname);
            elem.data(field, after);
            elem.find("." + field).removeClass(before.toString());
            elem.find("." + field).addClass(after.toString());
        }
    });
}

function setStateClass(elem, state) {
    elem.removeClass(stateNames.join(" "));
    elem.addClass(stateNames[state]);
}

function getClassName(host, service) {
    var name = host + "_" + service;
    name = name.replace(/ /g, "_")
    name = name.replace(/\./g, "_")
    return name
}

function init() {
    $('#filter').keyup(function() {
        var data = $('#data .panel');
        var re = new RegExp($(this).val(), "i");
        data.show().filter(function() {
            var name = $(this).find(".hostname").text();
            return !re.test(name);
        }).hide();
    });

    var url = new URL(window.location.href);
    var env = url.searchParams.get("env");
    if (env == null || env == undefined ) {
        env = "_"
    }
    loadEnv(env)
}

