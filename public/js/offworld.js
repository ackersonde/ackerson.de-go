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

  var url = "http://efa.mvv-muenchen.de/mvv/XSLT_TRIP_REQUEST2?&language=de"+
    "&anyObjFilter_origin=0&sessionID=0&itdTripDateTimeDepArr=dep&type_destination=any"+
    "&itdDateMonth="+month+"&itdTimeHour="+hour+"&anySigWhenPerfectNoOtherMatches=1"+
    "&locationServerActive=1&name_origin="+origin+"&itdDateDay="+day+"&type_origin=any"+
    "&name_destination="+destination+"&itdTimeMinute="+minute+"&Session=0&stateless=1"+
    "&SpEncId=0&itdDateYear="+year;

  var win = window.open(url, '_blank');
  win.focus();
}

(function($) {
  var id = 1;
  var greeting = "Welcome Off World (type help)";
  var drive_msg = "[[g;#FFFF00;]drive] <DESTINATION>: google directions from your location\r\n";
  var weather_msg = "[[g;#FFFF00;]weather]: show weather forecast\r\n";
  var whoami_msg = "[[g;#FFFF00;]whoami]: your browser info and IP address\r\n";
  var date_msg = "[[g;#FFFF00;]date]: my server date/time\r\n";
  var version_msg = "[[g;#FFFF00;]version]: build of this website\r\n";
  var sw_msg = "[[g;#FFFF00;]sw]: Schwabhausen weather \r\n";
  var clear_msg = "[[g;#FFFF00;]clear]: clear this terminal screen\r\n";
  var help = drive_msg + weather_msg + whoami_msg + date_msg + sw_msg + version_msg + clear_msg;

  // TODO 'wp' (write poetry) => window.open('https://draftin.com/api')
  // TODO 'poems' (see poems) => window.open('/poems')
  // => curl -u dan@ackerson.de:<pass> https://draftin.com/api/v1/documents.json
  // => filter by folder_id='20624' ('WRITINGS')

  $( "#drivePopup" ).draggable({ handle: "p.border" });
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

          case 'drive':
            showPopup(commands[0]);

            if (commands[1]) {
              document.getElementById('address').value = commands[1];
            } else {
              document.getElementById('address').value = 'Munich';
            }

            getDrivingDirections();
            break;

          case 'sw':
            schwabhausen_weather = '//www.wunderground.com/cgi-bin/findweather/getForecast?query=48.300000,11.350000&ID=IBAYERNS22';
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

  function getPosition() {
    if (currentLatLng === undefined || currentLat === undefined || currentLng === undefined) {
      navigator.geolocation.getCurrentPosition(function(position) {
        currentLatLng = new google.maps.LatLng(parseFloat(position.coords.latitude), parseFloat(position.coords.longitude));
        currentLat = currentLatLng.lat();
        currentLng = currentLatLng.lng();
        simpleAjaxCall('weather', {'lat':currentLat,'lng':currentLng});
        homeLocation = geocoder.geocode({'latLng': currentLatLng}, function(results, status) {
          if (status == google.maps.GeocoderStatus.OK) {
            if (results[0]) {
              // here we have to carefully scan through results and find city, country ('locality', 'political' in google geocode speak)
              for (i = 0; i < results.length; i++) {
                if (results[i].types[0] == 'locality' && results[i].types[1] == 'political') {
                  homeLocation = results[i].formatted_address;
                  break;
                }
              }
            }
          }
        });
      }, showGeoLocationError);
    } else {
      var currentLatLng = {"lat":currentLat,"lng":currentLng};
      // JSON.stringify(currentLatLng)
      simpleAjaxCall('weather', currentLatLng);
    }
  }

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
            var forecast = responseText['forecastday']['forecast']['simpleforecast']['forecastday'];
            var current = responseText['current']['current_observation']
            var weatherForecast = document.getElementById("forecastweather");
            weatherForecast.innerHTML = "";
            for(var i=0;i<forecast.length;i++){
                /*jshint multistr: true */
                weatherForecast.innerHTML += "\
                <div style='float:left;margin:10px;'>\
                    <span style='float:left;'>"+forecast[i]['date']['weekday_short']+",&nbsp;"+forecast[i]['date']['monthname']+" "+forecast[i]['date']['day']+"</span>\
                    <div style='float:left;clear:left;margin-right:5px;'>\
                        <span style='font-weight:bold;'>"+forecast[i]['low']['celsius']+"&nbsp;&#8451;</span>\
                        <img src='"+forecast[i]['icon_url']+"' width='44' height='44' alt='"+forecast[i]['conditions']+"'>\
                        <span style='font-weight:bold;'>"+forecast[i]['high']['celsius']+"&nbsp;&#8451;</span>\
                    </div>";
                    if (i+1 < forecast.length) {
                        weatherForecast.innerHTML += "<div style='border-right:1px solid lightgray;float:left;height:90px;'>&nbsp;</div>";
                    }
                weatherForecast.innerHTML += "</div>";
            }

            var weatherReport = document.getElementById("currentweather");
            weatherReport.innerHTML = "<span style='font-weight:bold;color:darkblue;'>Weather for "+homeLocation+"</span>\
                <div id='weatherreport'>\
                    <div style='float:left;margin-left:10px;'>\
                        <div>\
                            <a target='_blank' href='"+current['ob_url']+"'>\
                            <img src='"+current['icon_url']+"' width='44' height='44' alt='"+current['weather']+"'>\
                            </a>\
                        </div>\
                        <div style='margin-left:-10px;'>"+current['weather']+"</div>\
                    </div>\
                    <div style='float:left;margin-top:10px;margin-left:25px;text-align:left;'>\
                        <div style=''>Current \
                            <span style='font-weight:bold;'>"+current['temperature_string']+"</span>\
                        </div>\
                        <div>Feels Like\
                            <span style='font-weight:bold;'>"+current['feelslike_string']+"</span>\
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
