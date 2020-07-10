function get(url, callback) {
    fetch(url)
    .then(response => response.json())
    .then(json => {callback(json)})
}

export {get};