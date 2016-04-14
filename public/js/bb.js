function listDownloads(directory) {
  var xhttp;
  xhttp=new XMLHttpRequest();
  xhttp.onreadystatechange = function() {
    if (xhttp.readyState == 4 && xhttp.status == 200) {
      reloadFolder(xhttp);
    }
  };
  xhttp.open("GET", "/listDownloads?dir=" + directory, true);
  xhttp.send();
}

function reloadFolder(xhttp) {
  document.getElementById("downloadDirectory").innerHTML = xhttp.responseText;
}
