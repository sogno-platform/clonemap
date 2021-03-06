import * as request from "./request.js";

// initial website present
$(document).ready(function(){
    $("#navoverview").click(navoverview);
    $("#overview").click(navoverview);
    $("#inputMAS").change(inputMAS);
    $("#uploadMAS").click(uploadMAS);
    navoverview();
});

// callback for Overview button
function navoverview(){
    $(".modules").css("visibility","hidden");
    $(".nav-link").removeClass("active");
    $("#overview").addClass("active");
    //$("#headertitle").text("Overview");
    request.get('/api/overview',contentOverview);
    updateSidebar();
    // fetch('/api/ams/mas').then(response => response.json()).then(data => console.log(data));
}

// callback for mas buttons
function sidemas(){
    $(".modules").css("visibility","visible");
    $("#overview").removeClass("active");
    $(".masbutton").removeClass("active");
    $('#'+this.id).addClass("active");
    var masID = this.id.split("sidemas");
    
    contentAMS(parseInt(masID[1]));
}

// show sidebar and register callbacks for mas buttons
function updateSidebar(){
    fetch('/api/ams/mas')
    .then(response => response.json())
    // .then(data => JSON.parse(data))
    .then(mas => {
        $("#maslist").empty()
        if (mas === null) {
            return;
        }
        for (let i of mas) {
            $("#maslist").append("<li class='nav-item'><a href=\"#\" class=\"nav-link masbutton\" id=\"sidemas"+i.id.toString()+"\">MAS"+i.id.toString()+"</a></li>")
        }
        $(".masbutton").click(sidemas);
    })
}




// show content field for overview view
function contentOverview(mass) {
    clearContent();

    if (mass === null) {
        fetch("../snippets/empty-overview.html").then(response => {
            return response.text();
        }).then(emptyContent => {
            insertHtml("#contentContainer", emptyContent);
        }).catch(err => {
            console.log(err);
        })
        
        return;
    }

    $("#contentContainer").css("background-color", "rgb(240,240,240)");
    let startHtml = "<div class='contenttitle'><h2 class='lefttitle'>Overview</h2> \
    <button type='button' class='btn my-btn '\
    data-toggle='modal' data-target='#agencyModal'>\
    + New Agency</button></div> \
    <div class='container' id='tables'></div> ";
    insertHtml("#contentContainer", startHtml);
    let finalhtml = "<div class='row'>";
    fetch("../snippets/MAS-card.html").then(response => {
        return response.text();
    }).then(table => {
        let cnt = 0;
        for (let i of mass) {
            cnt++;
            let html = table;
            html = insertProperty(html,"count", cnt.toString());
            html = insertProperty(html,"name", i.config.name);
            html = insertProperty(html,"agent-per-agency", i.config.agentsperagency.toString());
            html = insertProperty(html,"DF", i.config.df.active.toString());
            html = insertProperty(html,"Logging", i.config.logger.active.toString());
            html = insertProperty(html,"MQTT", i.config.mqtt.active.toString());
            html = insertProperty(html,"Agents", i.numagents.toString());
            finalhtml += html;
        }
        finalhtml += "</div>";
        insertHtml("#tables", finalhtml);
    }).catch(error => {
        console.log(error);
    });

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
    request.post("/api/ams/mas", masconfig);
    navoverview();
}

// request info about mas and call function to show content
function contentAMS(masID){
    request.get('/api/ams/mas/'+masID.toString(),showAMSContent)
}

// show content field for ams and specified mas
function showAMSContent(masInfo) {
    clearContent();
    $(".modules").css("visibility","visible");
    fetch("../snippets/MAS.html").then(response => {
        return response.text();
    }).then( html => {
        html = insertProperty(html,"name", masInfo.config.name);
        html = insertProperty(html,"agent-per-agency", masInfo.config.agentsperagency.toString());
        html = insertProperty(html,"DF", masInfo.config.df.active.toString());
        html = insertProperty(html,"Logging", masInfo.config.logger.active.toString());
        html = insertProperty(html,"MQTT", masInfo.config.mqtt.active.toString());
        insertHtml("#contentContainer", html);
        let containersHtml = "<div class='row'>";
        for (let i of masInfo.imagegroups.instances) {
            containersHtml += "<div class='col-md-4 my-row'> <div class='card'>";
            containersHtml += ("<h5 class='card-header'>"+i.id.toString()+"</h5><div class='card-body'>");
            containersHtml += "<table class='table my-row'>";
            containersHtml += ("<tr><th>Image</th><th>"+i.config.image+"</th></tr>");
            containersHtml += ("<tr><th>Agencies:</th><th>"+i.agencies.counter.toString()+"</th></tr>");
            containersHtml += "</table> </div></div></div>"
        }
        containersHtml += "</div>";
        insertHtml("#containers", containersHtml);

        let agentsHtml = "<div class='row'>";
        for (let i of masInfo.agents.instances) {
            agentsHtml += "<div class='col-md-4 my-row'> <div class='card'>";
            containersHtml += ("<h5 class='card-header'>"+i.id.toString()+"</h5><div class='card-body'>");
            agentsHtml += "<table class='table my-row'>";
            agentsHtml += ("<tr><th>+Name:</th><th>"+i.spec.name+"</th></tr>");
            agentsHtml += ("<tr><th>Type:</th><th>"+i.spec.type+"</th></tr>");
            agentsHtml += ("<tr><th>Address:</th><th>"+i.address.agency+"</th></tr>");
            agentsHtml += "</table> </div></div></div>";
        }
        agentsHtml += "</div>"
        insertHtml("#agents", agentsHtml);
    }).catch(error => {
        console.log(error);
    });
}

// clear content field
function clearContent() {
    $(".contentcontainer").empty();
    
}

// Convenience function for inserting innerHTML for the selector
let insertHtml = function (selector, html) {
    let targetElem = document.querySelector(selector);
    targetElem.innerHTML = html;
  };

let insertProperty = function (string, propName, propValue) {
    let propToReplace = "{{" + propName + "}}";
    string = string.replace(new RegExp(propToReplace, "g"), propValue);
    return string;
  }

