import * as request from "./request.js";

$(document).ready(function(){

    $("#sideoverview").click(sideoverview);

    $("#sideplatform").click(sideplatform);

    updateSidebar();

});

function sideoverview(){
    $(".modules").hide();
    $("#headertitle").text("Overview");
    updateSidebar();
    contentOverview();
    // fetch('/api/ams/mas').then(response => response.json()).then(data => console.log(data));
}

function sideplatform(){
    $(".modules").hide();
    $("#headertitle").text("Platform");
    contentPlatform();
}

function sidemas(){
    $(".modules").show();
    var masID = this.id.split("sidemas")
    $("#headertitle").text("MAS"+masID[1]);
    contentAMS(parseInt(masID[1]))
}

function updateSidebar(){
    fetch('/api/ams/mas')
    .then(response => response.json())
    // .then(data => JSON.parse(data))
    .then(mas => {
        $("#maslist").empty()
        for (let i of mas) {
            $("#maslist").append("<li><a href=\"#\" class=\"masbutton\" id=\"sidemas"+i.id.toString()+"\">MAS"+i.id.toString()+"</a></li>")
        }
        $(".masbutton").click(sidemas);
    })
}

function contentOverview() {
    $(".contenttitle").replaceWith("<h2 class=\"contenttitle\">Overview</h2>");
    clearContent();
}

function contentPlatform() {
    $(".contenttitle").replaceWith("<h2 class=\"contenttitle\">Platform</h2>");
    clearContent();
}

function contentAMS(masID){
//     fetch('/api/ams/mas/'+masID.toString())
//     .then(response => response.json())
//     .then(masInfo => {
//         $(".contenttitle").replaceWith("<h2 class=\"contenttitle\">MAS"+masID.toString()+"</h2>");
//         contentMasInfo(masInfo);
//         console.log(masInfo);
//     })
    request.get('/api/ams/mas/'+masID.toString(),showAMSContent)
}

function showAMSContent(masInfo) {
    $(".contenttitle").replaceWith("<h2 class=\"contenttitle\">MAS"+masInfo.id.toString()+"</h2>");
    contentMasInfo(masInfo);
    console.log(masInfo);
}

function clearContent() {
    $(".content").empty();
}

function contentMasInfo(masInfo) {
    clearContent();
    $(".content").append("<table id=\"masinfoid\"></table>");
    $("#masinfoid").append("<tr><th>ID:</th><th>"+masInfo.id.toString()+"</th></tr>");
    $(".content").append("<hr>");
    $(".content").append("<table id=\"masinfoconfig\"></table>");
    $("#masinfoconfig").append("<tr><th>Config</th></tr>");
    $("#masinfoconfig").append("<tr><th></th><th>Name:</th><th>"+masInfo.config.name+"</th></tr>");
    $("#masinfoconfig").append("<tr><th></th><th>Agents per agency:</th><th>"+masInfo.config.agentsperagency.toString()+"</th></tr>");
    $("#masinfoconfig").append("<tr><th></th><th>DF:</th><th>"+masInfo.config.df.toString()+"</th></tr>");
    $("#masinfoconfig").append("<tr><th></th><th>Logging:</th><th>"+masInfo.config.logging.toString()+"</th></tr>");
    $("#masinfoconfig").append("<tr><th></th><th>MQTT:</th><th>"+masInfo.config.mqtt.toString()+"</th></tr>");
    $(".content").append("<hr>");
    $(".content").append("<table id=\"masinfocontainer\"></table>");
    $("#masinfocontainer").append("<tr><th>Containers</th></tr>");
    for (let i of masInfo.groups.instances) {
        $("#masinfocontainer").append("<tr><th></th><th>"+i.id.toString()+":</th><th>"+i.config.image+"</th></tr>");
        $("#masinfocontainer").append("<tr><th></th><th></th><th>Agencies:</th><th>"+i.agencies.counter.toString()+"</th></tr>");
    }
    $(".content").append("<hr>");
    $(".content").append("<table id=\"masinfoagents\"></table>");
    $("#masinfoagents").append("<tr><th>Agents</th></tr>");
    for (let i of masInfo.agents.instances) {
        $("#masinfoagents").append("<tr><th></th><th>"+i.id.toString()+":</th><th>Name:</th><th>"+i.spec.name+"</th></tr>");
        $("#masinfoagents").append("<tr><th></th><th></th><th>Type:</th><th>"+i.spec.type+"</th></tr>");
        $("#masinfoagents").append("<tr><th></th><th></th><th>Address:</th><th>"+i.address.agency+"</th></tr>");
    }
}