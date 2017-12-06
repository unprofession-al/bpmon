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
