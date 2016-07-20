function getParameterByName(name, url) {
    if (!url) url = window.location.href;
    name = name.replace(/[\[\]]/g, "\\$&");
    var regex = new RegExp("[?&]" + name + "(=([^&#]*)|&|#|$)"),
        results = regex.exec(url);
    if (!results) return null;
    if (!results[2]) return '';
    return decodeURIComponent(results[2].replace(/\+/g, " "));
}

var isCaller = getParameterByName('caller') !== null

function logError(error) {
    console.log(error.name + ': ' + error.message);
}

function SignalingChannel() {
    this.onmessage = function(evt) {
        console.log("Received message");
    }

    this.send = function(data) {
        console.log("Sending", data);
        this.ws.send(data)
    }

    this.connect = function() {
        console.log("Connecting");
        this.ws = new WebSocket("ws://localhost:10443/signaler", "signaler");
        this.ws.onopen = function () {
            this.send(JSON.stringify({
                'ID': isCaller ? 'caller' : 'callee'
            }));
        };
        this.ws.onmessage = this.onmessage;
    }

}

var signalingChannel = new SignalingChannel();
var configuration = {};
var pc;

function start() {
    pc = new RTCPeerConnection(configuration);

    // send any ice candidates to the other peer
    pc.onicecandidate = function (evt) {
        if (evt.candidate)
            signalingChannel.send(JSON.stringify({
                'candidate': evt.candidate,
                'dest': isCaller ? 'callee' : 'caller'
            }));
    };

    // let the 'negotiationneeded' event trigger offer generation
    pc.onnegotiationneeded = function () {
        console.log("Negotition needed")
        pc.createOffer(localDescCreated, logError);
    }

    // once remote stream arrives, show it in the remote video element
    pc.onaddstream = function (evt) {
        //remoteView.src = URL.createObjectURL(evt.stream);
    };

    // get a local stream, show it in a self-view and add it to be sent
    navigator.mediaDevices.getUserMedia({
        'audio': true,
        'video': false
    }).then(function (stream) {
        console.log("Add stream");
        //selfView.src = URL.createObjectURL(stream);
        pc.addStream(stream);
    }).catch(logError);
}

function localDescCreated(desc) {
    pc.setLocalDescription(desc, function () {
        signalingChannel.send(JSON.stringify({
            'sdp': pc.localDescription,
            'dest': isCaller ? 'callee' : 'caller'
        }));
    }, logError);
}

signalingChannel.onmessage = function (evt) {
    console.log('Received' + evt.data);
    if (!pc)
        start();
    var message = JSON.parse(evt.data);
    if (message.sdp) {
        console.log("Set remote description");
        pc.setRemoteDescription(new RTCSessionDescription(message.sdp), function () {
            // if we received an offer, we need to answer
            if (pc.remoteDescription.type == 'offer')
                pc.createAnswer(localDescCreated, logError);
        }, logError);
    } else {
        pc.addIceCandidate(new RTCIceCandidate(message.candidate));
    }
};

signalingChannel.connect();

if (isCaller) {
	start(true);
}
