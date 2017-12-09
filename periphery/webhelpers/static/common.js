function hideSpinner() {
        $("#loading").css("display","none");
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

function formatTimestamp(ts) {
	var options = {
    	day: "numeric", year: "numeric", month: "short",
        day: "numeric", hour: "2-digit", minute: "2-digit", second: "2-digit"
    };
    var date = new Date(ts);
    var dateString =  date.toLocaleTimeString("de-ch", options)
	return dateString
}
