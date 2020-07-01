$(document).ready(function(){

    $("#sideoverview").click(sideoverview);

    $("#sideplatform").click(sideplatform);

    updateSidebar();

});

function sideoverview(){
    $(".modules").hide();
    $("#headertitle").text("Overview");
    updateSidebar();
    // fetch('/api/ams/mas').then(response => response.json()).then(data => console.log(data));
}

function sideplatform(){
    $(".modules").hide();
    $("#headertitle").text("Platform");
}

function sidemas(){
    $(".modules").show();
    $("#headertitle").text(this.id);
}

function updateSidebar(){
    fetch('/api/ams/mas')
    .then(response => response.json())
    // .then(data => JSON.parse(data))
    .then(mas => {
        $("#maslist").empty()
        for (let i of mas) {
            console.log(i);
            $("#maslist").append("<li><a href=\"#\" class=\"masbutton\" id=\"sidemas"+i.id.toString()+"\">MAS"+i.id.toString()+"</a></li>")
        }
        $(".masbutton").click(sidemas);
    })
}