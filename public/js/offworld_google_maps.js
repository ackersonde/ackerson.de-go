var geocoder;
var map;
var green_marker;
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

  // default position to Munich
  currentLatLng = new google.maps.LatLng(48.1351, 11.5820);

  // Try HTML5 geolocation.
  if (navigator.geolocation) {
    getCurrentLocation();
  }
}

function getCurrentLocation() {
  navigator.geolocation.getCurrentPosition(function(position) {
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
