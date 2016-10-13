var geocoder;
var map;
var green_marker;
var currentPosition;
var currentLatLng;
var currentLat;
var currentLng;
var currentMarker;
var homeMarker;
var homeLocation;
var directionsDisplay;
var directionsService;
var ff_first_load = true;

function init_googlemaps() {
  geocoder = new google.maps.Geocoder();
  directionsDisplay = new google.maps.DirectionsRenderer();

  green_marker = new google.maps.MarkerImage(
    '/images/marker_greenA.png',
    new google.maps.Size(32, 32),   // size
    new google.maps.Point(0,0),     // origin
    new google.maps.Point(16, 32)   // anchor
  );

  getCurrentLocation();
}

function getCurrentLocation() {
  navigator.geolocation.getCurrentPosition(function(position) {
    currentPosition = position;
    currentLatLng = new google.maps.LatLng(parseFloat(position.coords.latitude), parseFloat(position.coords.longitude));
    homeLocation = geocoder.geocode({'latLng': currentLatLng}, function(results, status) {
      if (status == google.maps.GeocoderStatus.OK) {
        if (results[1]) {
          homeLocation = results[1].formatted_address;
          document.getElementById('home_location').innerHTML = 
            "<p style='font-size:14px;margin-left:10px;margin-bottom:0px;'><img style='height:24px;width:16px;vertical-align:middle;' src='/images/marker_greenA.png'>&nbsp;&nbsp;&nbsp;<b>" + homeLocation + "</b></p><hr>";
          return homeLocation;
        }
      }
    });
  }, showGeoLocationError);

  return currentLatLng;
}

function getDrivingDirections() {
  currentLatLng = getCurrentLocation();

  var address = document.getElementById('address').value;
  var mapOptions = {
    zoom: 8,
    mapTypeId: google.maps.MapTypeId.ROADMAP  // ROADMAP, SATELLITE, HYBRID, TERRAIN
  };

  map = new google.maps.Map(document.getElementById('map_canvas'), mapOptions);

  geocoder.geocode( { 'address': address }, function(results, status) {
    if (status == google.maps.GeocoderStatus.OK) {
      destinationLatLng = results[0].geometry.location; 
      if (ff_first_load) {
        alert('Driving to: ' + destinationLatLng);
        ff_first_load = false;
      }
         
      homeMarker = new google.maps.Marker({
          map: map,
          icon: green_marker,
          position: currentLatLng
      });
      directionsDisplay.setMap(map); // clear any past results

      var directionsRequest = {
        origin: homeMarker.position,
        destination: destinationLatLng,
        travelMode: google.maps.TravelMode.DRIVING // DRIVING, WALKING, BICYCLING, TRANSIT
      }
      homeMarker.setVisible(false); // remove dupe marker as directionsDisplay above sets home marker too

      directionsService = new google.maps.DirectionsService();
      directionsService.route(directionsRequest, function(result, status) {
        if (status == google.maps.DirectionsStatus.OK) {
          directionsDisplay.setDirections(result);

          var steps = result.routes[0].legs[0].steps
          var result = "";
          var totalTime = 0;
          var totalDist = 0;

          for ( var step in steps ) {
            result += "<p style='margin-left:10px;margin-top:0px;'><b>" + (parseInt(step) + 1) + ".</b> <span style='font-size:14px;'>" + steps[step].instructions + " (<span style='font-style:italic;font-size:12px;color:blue;'>" + steps[step].distance.text + " - " + steps[step].duration.text + "</span>)</span></p>";

            totalTime += steps[step].duration.value;
            totalDist += steps[step].distance.value;
          }

          hours = parseInt( totalTime / 3600 ) % 24;
          minutes = parseInt( totalTime / 60 ) % 60;
          totalTime = hours + "hr " + (minutes < 10 ? "0" + minutes : minutes) + "mins";

          document.getElementById('directions').innerHTML = result;
          document.getElementById('home_location').innerHTML = 
            "<p style='font-size:14px;margin-left:10px;margin-bottom:0px;'><img style='height:24px;width:16px;vertical-align:middle;' src='/images/marker_greenA.png'>&nbsp;&nbsp;&nbsp;<b>" + homeLocation + "</b></p>" + 
            "<p style='font-size:14px;margin-left:10px;margin-bottom:0px;'><img style='height:24px;width:16px;vertical-align:middle;' src='/images/marker_greenB.png'>&nbsp;&nbsp;&nbsp;<b>" + document.getElementById('address').value + "</b> (<span style='font-size:14px;font-style:italic;font-weight:bold;color:blue;'>" + totalDist/1000 + " km - " + totalTime + "</span>)</p><hr>";
        }
      });
    } else {
      alert('Geocode was not successful for the following reason: ' + status);
    }
  });

  return currentLatLng;
}

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
