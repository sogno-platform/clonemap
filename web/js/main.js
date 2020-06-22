$(document).ready(function(){

    $("#sideoverview").click(sideoverview);

    $("#sideplatform").click(sideplatform);

    $(".masbutton").click(sidemas);

});

function sideoverview(){
    $(".modules").hide()
    $("#headertitle").text("Overview")
}

function sideplatform(){
    $(".modules").hide()
    $("#headertitle").text("Platform")
}

function sidemas(){
    $(".modules").show()
    $("#headertitle").text(this.id)
}