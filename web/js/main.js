$(document).ready(function(){

    $("#sideoverview").click(sideoverview);

    $("#sideplatform").click(sideplatform);

    $(".masbutton").click(sidemas);

});

function sideoverview(){
    $(".modules").hide();
    $("#headertitle").text("Overview");
    fetch('/api/ams/mas').then(response => response.json()).then(data => console.log(data));
}

function sideplatform(){
    $(".modules").hide();
    $("#headertitle").text("Platform");
}

function sidemas(){
    $(".modules").show();
    $("#headertitle").text(this.id);
}