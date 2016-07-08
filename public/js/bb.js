function getMLBFormattedDate(date) {
  var month = "%2fmonth_" + ("0" + (date.getMonth() + 1)).slice(-2);
  var day = "%2fday_" + ("0" + date.getDate()).slice(-2)
  return "year_" + date.getFullYear() + month + day;
}

function notifyAjaxCall(renderImage) {
    var ballIcon = document.getElementById('ball');
    if (ballIcon != null) {
      ballIcon.src = 'images/' + renderImage;
    }
}

function prepareDate(currentDate) {
  var searchDate;

  // current GET https://ackerson.de/bb?date1=year_2016%2fmonth_04%2fday_26&offset=1
  if (currentDate == "") {
    var today = new Date();
    searchDate = today.setDate(today.getDate() - 1);
  } else {
      var searchDateString = document.getElementById('searchDate').value;

      // take the 'Wed, Apr 27 2016' string above and convert to Date object
      searchDate = new Date(Date.parse(searchDateString)) // millis

  }
  return getMLBFormattedDate(searchDate)
}

function initializeDatePicker(clazz) {
  $(function() {
    $( clazz ).datepicker({
      onSelect: function(value, date)
        {
          notifyAjaxCall('spinningBall.gif');
          $.ajax({
            type:"GET",
            url: "/bbAjaxDay?date1=" + prepareDate(value) + "&offset=" + 0,
            success: function(result)
            {
              notifyAjaxCall('pokemon.jpg');
              document.getElementById("responseBB").innerHTML=result;
              initializeDatePicker(clazz);
            }
          });
        },
      dateFormat: "D, M d yy",
      showButtonPanel: true,
      showAnim: 'slideDown',
      autoSize: true,
      showOn: "both",
      buttonImage: "images/calendar.gif",
      buttonImageOnly: true,
      buttonText: "Select date"
    });
  });
}

function fetchGames(date1, offset) {
    notifyAjaxCall('spinningBall.gif');
    $.ajax({
      type:"GET",
      url: "/bbAjaxDay?date1=" + date1 + "&offset=" + offset,
      success: function(result)
      {
        notifyAjaxCall('pokemon.jpg');
        document.getElementById("responseBB").innerHTML=result;
        initializeDatePicker('#searchDate');
      }
    });
}

fetchLast4WeeksForFavTeam(favTeamID) {
  $.ajax({
    type:"GET",
    url: "/clockCheck?panel=panel1",
    success: function(result)
    {
      document.getElementById("panel1").innerHTML=result;
    }
  });
  $.ajax({
    type:"GET",
    url: "/clockCheck?panel=panel2",
    success: function(result)
    {
      document.getElementById("panel2").innerHTML=result;
    }
  });
  $.ajax({
    type:"GET",
    url: "/clockCheck?panel=panel3",
    success: function(result)
    {
      document.getElementById("panel3").innerHTML=result;
    }
  });
  $.ajax({
    type:"GET",
    url: "/clockCheck?panel=panel4",
    success: function(result)
    {
      document.getElementById("panel4").innerHTML=result;
    }
  });
}
