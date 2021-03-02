import * as request from "./request.js";

// initial website present
$(document).ready(function(){
    $("#sideoverview").click(sideoverview);
    updateSidebar();
    sideoverview();
});

// callback for Overview button
function sideoverview(){
    $(".modules").hide();
    $("#headertitle").text("Overview");
    updateSidebar();
    request.get('/api/overview',contentOverview)
    // fetch('/api/ams/mas').then(response => response.json()).then(data => console.log(data));
}

// callback for mas buttons
function sidemas(){
    $(".modules").show();
    var masID = this.id.split("sidemas")
    $("#headertitle").text("MAS"+masID[1]);
    contentAMS(parseInt(masID[1]))
}

// show sidebar and register callbacks for mas buttons
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

// show content field for overview view
function contentOverview(mass) {
    $(".contenttitle").replaceWith("<h2 class=\"contenttitle\">Overview</h2>");
    clearContent();
    $(".content").append("<div class=\"contentbox\" id=\"startbox\"></div>")
    $("#startbox").append("<table id=\"startmas\"></table>");
    $("#startmas").append("<tr><th><h3>Start new MAS:</h3></th></tr>")
    $("#startbox").append("<hr>");
    $("#startbox").append("<tr><th><input type=\"file\" id=\"inputMAS\" value=\"Import\" accept=\".json\"/></th><th><button id=\"uploadMAS\">Upload</button></th></tr>")
    $("#inputMAS").change(inputMAS);
    $("#uploadMAS").click(uploadMAS);
    $("#startbox").append("<tr><th><textarea id=\"newMAS\"></textarea></th></tr>")
    $(".content").append("<div class=\"contentbox\" id=\"massbox\"></div>")
    $("#massbox").append("<table id=\"mass\"></table>");
    $("#mass").append("<tr><th><h3>MASs:</h3></th><th>"+mass.length.toString()+"</th></tr>");
    $("#massbox").append("<hr>");
    for (let i of mass) {
        let masID = "MAS"+i.id.toString()
        $("#massbox").append("<table id=\""+masID+"\"></table>");
        $("#"+masID).append("<tr><th>"+masID+"</th></tr>");
        $("#"+masID).append("<tr><th></th><th>Name:</th><th>"+i.config.name+"</th></tr>");
        $("#"+masID).append("<tr><th></th><th>Agents per agency:</th><th>"+i.config.agentsperagency.toString()+"</th></tr>");
        $("#"+masID).append("<tr><th></th><th>DF:</th><th>"+i.config.df.active.toString()+"</th></tr>");
        $("#"+masID).append("<tr><th></th><th>Logging:</th><th>"+i.config.logger.active.toString()+"</th></tr>");
        $("#"+masID).append("<tr><th></th><th>MQTT:</th><th>"+i.config.mqtt.active.toString()+"</th></tr>");
        $("#"+masID).append("<tr><th></th><th>Agents:</th><th>"+i.numagents.toString()+"</th></tr>");
    }
}

// callback for inputMAS input button
function inputMAS() {
    let files = this.files;
    if (files.length <= 0) {
        return false;
    }

    let fr = new FileReader();
    fr.onload = function(e) { 
        var result = JSON.parse(e.target.result);
        var formatted = JSON.stringify(result, null, 2);
            $("#newMAS").val(formatted);
    }

    fr.readAsText(files.item(0));
}

// callback for uploadMAS button
function uploadMAS() {
    let masconfig = $("#newMAS").val();
    request.post("/api/ams/mas", masconfig)
    sideoverview()
}

// request info about mas and call function to show content
function contentAMS(masID){
    request.get('/api/ams/mas/'+masID.toString(),showAMSContent)
}

// show content field for ams and specified mas
function showAMSContent(masInfo) {
    $(".contenttitle").replaceWith("<h2 class=\"contenttitle\">MAS"+masInfo.id.toString()+"</h2>");
    contentMasInfo(masInfo);
    console.log(masInfo);
}

// clear content field
function clearContent() {
    $(".content").empty();
}

// show content field for mas info
function contentMasInfo(masInfo) {
    clearContent();
    $(".content").append("<table id=\"masinfoid\"></table>");
    $("#masinfoid").append("<tr><th>ID:</th><th>"+masInfo.id.toString()+"</th></tr>");
    $(".content").append("<hr>");
    $(".content").append("<table id=\"masinfoconfig\"></table>");
    $("#masinfoconfig").append("<tr><th>Config</th></tr>");
    $("#masinfoconfig").append("<tr><th></th><th>Name:</th><th>"+masInfo.config.name+"</th></tr>");
    $("#masinfoconfig").append("<tr><th></th><th>Agents per agency:</th><th>"+masInfo.config.agentsperagency.toString()+"</th></tr>");
    $("#masinfoconfig").append("<tr><th></th><th>DF:</th><th>"+masInfo.config.df.active.toString()+"</th></tr>");
    $("#masinfoconfig").append("<tr><th></th><th>Logging:</th><th>"+masInfo.config.logger.active.toString()+"</th></tr>");
    $("#masinfoconfig").append("<tr><th></th><th>MQTT:</th><th>"+masInfo.config.mqtt.active.toString()+"</th></tr>");
    $(".content").append("<hr>");
    $(".content").append("<table id=\"masinfocontainer\"></table>");
    $("#masinfocontainer").append("<tr><th>Containers</th></tr>");
    for (let i of masInfo.imagegroups.instances) {
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