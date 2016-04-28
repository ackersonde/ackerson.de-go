function getMLBFormattedDate(date) {
  var month = "%2fmonth_" + ("0" + (date.getMonth() + 1)).slice(-2);
  var day = "%2fday_" + ("0" + date.getDate()).slice(-2)
  return "year_" + date.getFullYear() + month + day;
}

function fetchGames(currentDate, offset) {
  var date1;
  var offset;

  // current GET https://ackerson.de/bb?date1=year_2016%2fmonth_04%2fday_26&offset=1
  if (currentDate == "") {
    var today = new Date();
    var yesterday = today.setDate(today.getDate() - 1);
    date1 = getMLBFormattedDate(yesterday)
  } else {
      searchDateString = document.getElementById('searchDate').innerText;

      // take the 'Wed, Apr 27 2016' string above and convert to Date object
      searchDate = new Date(Date.parse(searchDateString)) // millis

      date1 = getMLBFormattedDate(searchDate)
  }

  if (window.XMLHttpRequest) {
    xmlhttp=new XMLHttpRequest();
  }

  xmlhttp.onreadystatechange=function() {
    if (xmlhttp.readyState == 4 && xmlhttp.status == 200) {
      document.getElementById("responseBB").innerHTML=xmlhttp.responseText;
    }
  }

  xmlhttp.open("GET", "/bbDay?date1=" + date1 + "&offset=" + offset, true);
  xmlhttp.send();
}
