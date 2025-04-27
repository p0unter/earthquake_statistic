fetch('/eq-rows')
    .then(res => res.text())
    .then(htmlRows => {
        document.getElementById('earthquake-tbody').innerHTML = htmlRows;
    })
    .catch(err => {
        document.getElementById('earthquake-tbody').innerHTML = '<tr><td colspan="8">Error: bad request.</td></tr>';
        console.log('Fetch error: ', err);
    });
