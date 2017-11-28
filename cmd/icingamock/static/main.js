var stateNames = [ "ok", "warn", "critical", "unknown" ];
var wsopen = false;
var connection = new WebSocket('ws://' + window.location.host + '/ws');

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
    var after = ( Number(before) + 1 ) % 4;
    var data = {
        env: env,
        host: host,
        service: service,
        attrs: {
            state: after
        },
    };
    connection.send(JSON.stringify(data));
    /*
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
    */
}

function toggleBool(env, host, service, field) {
    var classname = getClassName(host, service);
    var before = $("#" + classname).data(field);
    var after = ( before != true );
    var data = {
        env: env,
        host: host,
        service: service,
        attrs: {},
    };
    data.attrs[field] = after;
    connection.send(JSON.stringify(data));
    /*
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
    */
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
    $(document).on( 'keyup', '#filter', function() {
        var data = $('#data .panel');
        console.log($(this).text())
        var re = new RegExp($(this).text(), "i");
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

connection.onopen = function () {
    wsopen = true;
};

connection.onerror = function (error) {
    wsopen = false;
    $("<div class='notification panel'>WebSocket closed...</div>").hide().appendTo("#data").fadeIn();
};

connection.onmessage = function (e) {
    var instr = JSON.parse(e.data);
    var classname = getClassName(instr.host, instr.service);
    var elem = $("#" + classname);

    for (var attr in instr.attrs) {
        switch(attr) {
            case "state":
                setState(elem, instr.attrs[attr]);
                break;
            case "downtime":
                setBool(elem, attr, instr.attrs[attr]);
                break;
            case "acknowledgement":
                setBool(elem, attr, instr.attrs[attr]);
                break;
            default:
                var message = "Unknown instruction recieved: <br><pre>" + e.data + "</pre>";
                notify(message);
        }
    }
};

function setState(elem, state) {
    var before = elem.data("check-state");
    elem.data("check-state", state);
    var subelem = elem.find(".state");
    subelem.text(stateNames[state]);
    setStateClass(subelem, state);
}

function setBool(elem, field, state) {
    var before = elem.data(field);
    elem.data(field, state)
    elem.find("." + field).removeClass(before.toString());
    elem.find("." + field).addClass(state.toString());
}

function notify(message) {
    var id = makeid();
    var out = "<div class='notification panel' id='"+id+"'>"+message+"</div>";
    $("#data").prepend(out);
    $('#'+id).delay(5000).fadeOut('slow');
}

function makeid() {
    var text = "";
    var possible = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";

    for (var i = 0; i < 5; i++)
        text += possible.charAt(Math.floor(Math.random() * possible.length));

    return text;
}
