function get(url, callback) {
    fetch(url)
    .then(response => response.json())
    .then(json => {callback(json)})
}

function post(url, data) {
    fetch(url, {
    	method: 'post',
    	headers: {'Content-Type' : 'application/json'},
    	body: data
    })
    .then ( (text) => {
        // log response text 
        console.log (text);
    })
    .catch ((error) => {
        console.log ("Error: ", error)
    })
}

export {get, post};