// Ionic Starter App


var $rootScope = {};
var message = "Ready";
var lastMessage = "";
var statusStr="";
var peripheral = null;
var chargingRE = new RegExp("Charging", "i");
var keyboardRE = new RegExp("keyboard", "i");
var hrRE = new RegExp("hr", "i");

   function scan() {

        var foundHeartRateMonitor = false;

        function onScan(peripheral) {
            // this is demo code, assume there is only one heart rate monitor
            status("Found " + JSON.stringify(peripheral), ".  Connecting");
            foundHeartRateMonitor = true;

	    if (peripheral.name != "ID115") {return;}
        if(typeof ble  !== 'undefined') {
            ble.connect(peripheral.id, onConnect, onDisconnect);
	}
        }

        function scanFailure(reason) {
            status("BLE Scan Failed because "+ reason);
        }

        //ble.scan([heartRate.service], 5, onScan, scanFailure);
        if(typeof ble  !== 'undefined') {
        status("Scanning for Heart Rate Monitor");
        ble.scan([], 5, onScan, scanFailure);
	}

        setTimeout(function() {
            if (!foundHeartRateMonitor) {
                status("Did not find a device at all");
            }
        }, 5000);
    };
    function parseHexString(str) {
    var result = [];
    while (str.length >= 2) {
        result.push(parseInt(str.substring(0, 2), 16));
        str = str.substring(2, str.length);
    }
    return result;
}

function hex2bin(bytes, hex){
    for(var i=0; i< hex.length-1; i+=2) {
        bytes.push(parseInt(hex.substr(i, 2), 16));
    }
    return String.fromCharCode.apply(String, bytes);
}
        function stringToAsciiByteArray(str)
                {
                    var bytes = [];
                   for (var i = 0; i < str.length; ++i)
                   {
                       var charCode = str.charCodeAt(i);
                      if (charCode > 0xFF)  // char > 1 byte since charCodeAt returns the UTF-16 value
                      {
                          throw new Error('Character ' + String.fromCharCode(charCode) + ' can\'t be represented by a US-ASCII byte.');
                      }
                       bytes.push(charCode);
                   }
                    return bytes;
                }
    function onConnect(p) {
        if (p.name != "ID115") {return;}
        if(typeof notificationListener  !== 'undefined') {
                notificationListener.listen(function(n){
                 status("Received notification " + JSON.stringify(n) );
		if (chargingRE.exec(n.title)) { console.log("Not sending charging message"); return; }
                 message = n.title;
                if (n.text != "") {
                        message = n.text;
                }
                 sendNotif();
               }, function(e){
                 status("Notification Error " + e);
               });
        } else {
		 alert("Notification module failed!");
	}
        status("Added notification listener");

        status("Found watch");
        peripheral = p;
        setInterval(sendNotif, 5000);
        status("Connected to " + peripheral.id);
        sendNotif();
        return;
    }

function padRight(text, pad, maxLen) {
	if (text.length >= maxLen) { return text; }
	return padRight(text+pad, pad, maxLen);
}
    function sendNotif() {
        if(typeof peripheral  === 'undefined') { return; }
        if(peripheral  == null) { return; }
        if (message == lastMessage) { return; }
        if (chargingRE.exec(message)) { console.log("Not sending charging message"); return; }
        if (keyboardRE.exec(message)) { return; }
        if (hrRE.exec(message)) { return; }
        status("Sending notification message '" + message + "' to " + peripheral.id);
        var message_length = message.length;
        if (message_length>12) { message = message.substr(0,11); }
        lastMessage = message;

	var myMessage = padRight(message, " ", 12);
        message_length = myMessage.length;
        var message_bytes = stringToAsciiByteArray(myMessage);
        var bytes = [5, 3, 1, 1, 1,message_length, 0, 8];
        var packet_bytes = bytes.concat(message_bytes);
        var packet = new Uint8Array(packet_bytes);
        //status("Sending " + packet_bytes);
        ble.write(peripheral.id, "0af0", "0af6", packet.buffer, writeSuccess, writeFailure);
    }
   function onDisconnect(reason) {
        status("Disconnected " + reason);
        beatsPerMinute.innerHTML = "...";
        status("Disconnected");
    }
    function writeSuccess() {
        //status("Write successful");
    }
    function writeFailure(err) {
        status("Write to device failed because: " + err + ", restarting scan");
    }
    function onData(buffer) {
        // assuming heart rate measurement is Uint8 format, real code should check the flags 
        // See the characteristic specs http://goo.gl/N7S5ZS
        var data = new Uint8Array(buffer);
        beatsPerMinute.innerHTML = data[1];
    }
    function onError(reason) {
        alert("There was an error " + reason);
    }
    function status(message) {
	if ($rootScope.settings.enableFriends) {
		statusStr = " | <p> " + message + statusStr;
	} else {
		statusStr = message;
	}
        console.log(statusStr);
        statusDiv.innerHTML = statusStr;
    }


// angular.module is a global place for creating, registering and retrieving Angular modules
// 'starter' is the name of this angular module example (also set in a <body> attribute in index.html)
// the 2nd parameter is an array of 'requires'
// 'starter.services' is found in services.js
// 'starter.controllers' is found in controllers.js
angular.module('starter', ['ionic', 'starter.controllers', 'starter.services'])

.run(function($ionicPlatform) {
  $ionicPlatform.ready(function() {
    // Hide the accessory bar by default (remove this to show the accessory bar above the keyboard
    // for form inputs)
    if (window.cordova && window.cordova.plugins && window.cordova.plugins.Keyboard) {
      cordova.plugins.Keyboard.hideKeyboardAccessoryBar(true);
      cordova.plugins.Keyboard.disableScroll(true);

    }
    if (window.StatusBar) {
      // org.apache.cordova.statusbar required
      StatusBar.styleDefault();
    }
    scan();
  });
})

.config(function($stateProvider, $urlRouterProvider) {

  // Ionic uses AngularUI Router which uses the concept of states
  // Learn more here: https://github.com/angular-ui/ui-router
  // Set up the various states which the app can be in.
  // Each state's controller can be found in controllers.js
  $stateProvider

  // setup an abstract state for the tabs directive
    .state('tab', {
    url: '/tab',
    abstract: true,
    templateUrl: 'templates/tabs.html'
  })

  // Each tab has its own nav history stack:

  .state('tab.dash', {
    url: '/dash',
    views: {
      'tab-dash': {
        templateUrl: 'templates/tab-dash.html',
        controller: 'DashCtrl'
      }
    }
  })

  .state('tab.chats', {
      url: '/chats',
      views: {
        'tab-chats': {
          templateUrl: 'templates/tab-chats.html',
          controller: 'ChatsCtrl'
        }
      }
    })
    .state('tab.chat-detail', {
      url: '/chats/:chatId',
      views: {
        'tab-chats': {
          templateUrl: 'templates/chat-detail.html',
          controller: 'ChatDetailCtrl'
        }
      }
    })

  .state('tab.account', {
    url: '/account',
    views: {
      'tab-account': {
        templateUrl: 'templates/tab-account.html',
        controller: 'AccountCtrl'
      }
    }
  });

  // if none of the above states are matched, use this as the fallback
  $urlRouterProvider.otherwise('/tab/dash');

});
