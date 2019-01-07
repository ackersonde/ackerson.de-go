function popupClose(id) {
  $('#' + id).hide();
}

function showPopup(id) {
  $('#' + id + 'Popup').fadeIn('slow');
}

function mvvRoute(origin, destination) {
  var d = new Date();
  var year = d.getFullYear();
  var month = d.getMonth() + 1;
  var day = d.getDate();
  var hour = d.getHours();
  var minute = d.getMinutes();

  var url = "https://efa.mvv-muenchen.de/index.html?&language=en"+
    "&anyObjFilter_origin=0&sessionID=0&itdTripDateTimeDepArr=dep&type_destination=any"+
    "&itdDateMonth="+month+"&itdTimeHour="+hour+"&anySigWhenPerfectNoOtherMatches=1"+
    "&locationServerActive=1&name_origin="+origin+"&itdDateDay="+day+"&type_origin=any"+
    "&name_destination="+destination+"&itdTimeMinute="+minute+"&Session=0&stateless=1"+
    "&SpEncId=0&itdDateYear="+year+"#trip@origdest";

  var win = window.open(url, '_blank');
  win.focus();
}

(function($) {
  var id = 1;
  var greeting = "Welcome Off World (type help)";
  var weather_msg = "[[g;#FFFF00;]weather]: show weather forecast\r\n";
  var whoami_msg = "[[g;#FFFF00;]whoami]: your browser info and IP address\r\n";
  var date_msg = "[[g;#FFFF00;]date]: my server date/time\r\n";
  var version_msg = "[[g;#FFFF00;]version]: build of this website\r\n";
  var sw_msg = "[[g;#FFFF00;]sw]: Schwabhausen weather \r\n";
  var clear_msg = "[[g;#FFFF00;]clear]: clear this terminal screen\r\n";
  var help = weather_msg + whoami_msg + date_msg + sw_msg + version_msg + clear_msg;

  $( "#weatherPopup" ).draggable({ handle: "p.border" });

  var anim = false;
  function typedPrompt(term, message, delay) {
    anim = true;
    var c = 0;
    var interval = setInterval(function() {
      term.insert(message[c++]);
      if (c == message.length) {
        clearInterval(interval);
        setTimeout(function() {
          anim = false;
        }, delay);
      }
    }, delay);
  }

  term = $('#term1').terminal(function(command, term) {
    var commands = command.split(' ');
    if (commands.length > 0) {
      try {
        switch (commands[0]) {
          case 'help':
            term.echo(help);
            break;

          case 'date':
          case 'version':
          case 'whoami':
            simpleAjaxCall(command, "query-param");
            break;

          case 'sw':
            schwabhausen_weather = '//darksky.net/forecast/48.3,11.357/ca24/en#week';
            window.open(schwabhausen_weather);
            break;

          case 'weather':
            getPosition();
            break;

          default:
            /*jslint es5: true */
            var result = window.eval(command);
            if (result !== undefined) {
                term.echo(result);
            }
            break;
        }
      } catch(e) {
        term.echo("[[guib;#FFFF00;]" + e + "] (try `help`)");
      }
    } else {
      term.echo('');
    }
  }, {
    greetings: greeting,
    name: 'term1',
    enabled: false,
    prompt: 'dan@ackerson.de:~ $ ',
    onInit: function(term) {
      //typedPrompt(term, 'help', 250);
    },
    onClear: function(term) {
      term.echo(greeting);
    },
    keydown: function(e) {
      //disable keyboard when animating
      if (anim) {
        return false;
      }
    }
  });

  function showGeoLocationError(error) {
    switch(error.code) {
      case error.PERMISSION_DENIED:
        alert("User denied the request for Geolocation.");
        break;
      case error.POSITION_UNAVAILABLE:
        alert("Location information is unavailable.");
        break;
      case error.TIMEOUT:
        alert("The request to get user location timed out.");
        break;
      case error.UNKNOWN_ERROR:
        alert("An unknown error occurred.");
        break;
    }
  }

   var currentLat;
   var currentLng;

  function getPosition() {
    navigator.geolocation.getCurrentPosition(function(position) {
      currentLat = parseFloat(position.coords.latitude);
      currentLng = parseFloat(position.coords.longitude);
      simpleAjaxCall('weather', {'lat':currentLat,'lng':currentLng});
    }, showGeoLocationError);
  }

  var today = new Date();
  var weekday = new Array(7);
  weekday[0] =  "Sun";
  weekday[1] = "Mon";
  weekday[2] = "Tue";
  weekday[3] = "Wed";
  weekday[4] = "Thu";
  weekday[5] = "Fri";
  weekday[6] = "Sat";
  var month = new Array();
  month[0] = "Jan";
  month[1] = "Feb";
  month[2] = "Mar";
  month[3] = "Apr";
  month[4] = "May";
  month[5] = "Jun";
  month[6] = "Jul";
  month[7] = "Aug";
  month[8] = "Sep";
  month[9] = "Oct";
  month[10] = "Nov";
  month[11] = "Dec";

  /* simple ajax call where typed cmd string is SAME as remote URI AND data set */
  function simpleAjaxCall(command, query_param) {
    term.pause();

    //$.jrpc is helper function which creates json-rpc request
    $.jrpc(command,                         // uri
      id++,
      query_param,
      'post',
      function(data) {
        term.resume();
        if (data.error) {
          term.error(data.error.message);
        } else {
          var responseText = jQuery.parseJSON(data.responseText);
          if (command == 'version') {
            term.echo("[[g;#FFFF00;]ackerson.de build " + responseText['build'] + "]")
            window.open(responseText['version']);
          }
          else if (command == 'weather') {
            showPopup(command);
            var darkSkyIconsURL = "https://darksky.net/images/weather-icons/";
            var forecast = responseText['forecastday']['data'];
            var current = responseText['current'];
            var units = responseText['units'] == 'us' ? '&#8457;' : '&#8451;';
            var weatherForecast = document.getElementById("forecastweather");
            weatherForecast.innerHTML = "";
            for(var i=1;i<5;i++){
                var sslImage = darkSkyIconsURL + forecast[i]['icon'] + ".png";
                var dayWeather = new Date();
                dayWeather.setDate(today.getDate() + i);
                var dateWeather = weekday[dayWeather.getDay()]+",&nbsp;"+
                  month[dayWeather.getMonth()]+"&nbsp;"+dayWeather.getDate();

                weatherForecast.innerHTML += "\
                <div style='float:left;margin:10px;'>\
                    <span style='float:left;'>"+dateWeather+"</span>\
                    <div style='float:left;clear:left;margin-right:5px;'>\
                        <span style='font-weight:bold;'>"+Math.round(forecast[i]['temperatureMin'])+"&nbsp;"+units+"</span>\
                        <img src='"+sslImage+"' width='44' height='44' style='background:white;'>\
                        <span style='font-weight:bold;'>"+Math.round(forecast[i]['temperatureMax'])+"&nbsp;"+units+"</span>\
                    </div>";
                    if (i+1 < 5) {
                        weatherForecast.innerHTML += "<div style='border-right:1px solid lightgray;float:left;height:90px;'>&nbsp;</div>";
                    }
                weatherForecast.innerHTML += "</div>";
            }

            var weatherReport = document.getElementById("currentweather");
            var sslCurrentImage = darkSkyIconsURL + current['icon'] + ".png";
            var yourDarkSkyWeather = "https://darksky.net/forecast/" +
              currentLat + "," + currentLng + "/auto24";

            weatherReport.innerHTML = "<span style='font-weight:bold;color:white;'>Your weather</span>\
                <div id='weatherreport'>\
                    <div style='float:left;margin-left:10px;'>\
                        <div>\
                            <a target='_blank' href='"+yourDarkSkyWeather+"'>\
                            <img src='"+sslCurrentImage+"' width='44' height='44' style='background:white;'>\
                            </a>\
                        </div>\
                        <div style='margin-left:-10px;'>"+current['summary']+"</div>\
                    </div>\
                    <div style='float:left;margin-top:10px;margin-left:25px;text-align:left;'>\
                        <div style=''>Current \
                            <span style='font-weight:bold;'>"+
                            Math.round(current['temperature'])+"&nbsp;"+units+"</span>\
                        </div>\
                        <div>Feels Like\
                            <span style='font-weight:bold;'>"+
                            Math.round(current['apparentTemperature'])+"&nbsp;"+units+"</span>\
                        </div>\
                    </div>\
                </div>\
                ";
          } else term.echo(responseText[command]);       // data set
        }
      },
      function(xhr, status, error) {
        term.error('[AJAX] ' + status + ' - Server reponse is: \n' +
                    xhr.responseText);
        term.resume();
      }
    ); // rpc call
  }

  term.mouseout(function() {
    term.focus(false);
  });
  term.mouseover(function() {
    term.focus(true);
  });
})(jQuery);
