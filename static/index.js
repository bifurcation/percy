// https://blog.mozilla.org/webrtc/signaling-with-rtcsimpleconnection/

let gUMConfig = { "audio": false, "video": true };
const IP_PORT_REGEX = /\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\s+\d+/;
const RELAY_IP = "RELAY_IP_FROM_GO_SERVER";
const RELAY_PORT = "RELAY_PORT_FROM_GO_SERVER";
const IPV6_REGEX = RegExp('\:[0-9a-f]*\:[0-9a-fA-f]*','g');
const TCP_REGEX = RegExp('.*tcptype.*','g');

const ANSWER_SDP = {type: "answer", sdp: "v=0\r\no=percy0.1 2633292546686233323 0 IN IP4 0.0.0.0\r\ns=-\r\nt=0 0\r\na=fingerprint:sha-256 AA:AA:AA:AA:AA:AA:AA:AA:AA:AA:AA:AA:AA:AA:AA:AA:AA:AA:AA:AA:AA:AA:AA:AA:AA:AA:AA:AA:AA:AA:AA:AA\r\na=group:BUNDLE sdparta_0\r\na=ice-options:trickle\r\na=msid-semantic:WMS *\r\nm=video 9 UDP/TLS/RTP/SAVPF 120\r\nc=IN IP4 0.0.0.0\r\na=recvonly\r\na=extmap:3 urn:ietf:params:rtp-hdrext:sdes:mid\r\na=extmap:4 http://www.webrtc.org/experiments/rtp-hdrext/abs-send-time\r\na=extmap:5 urn:ietf:params:rtp-hdrext:toffset\r\na=fmtp:120 max-fs=12288;max-fr=60\r\na=ice-pwd:abcdefabcdefabcdefabcdefabcdefab\r\na=ice-ufrag:fedcbafe\r\na=mid:sdparta_0\r\na=rtcp-fb:120 nack\r\na=rtcp-fb:120 nack pli\r\na=rtcp-fb:120 ccm fir\r\na=rtcp-fb:120 goog-remb\r\na=rtcp-mux\r\na=rtpmap:120 VP8/90000\r\na=setup:passive\r\n"};

// Handy element access
let page = {
  get run() { return document.getElementById("run"); },
  get offer() { return document.getElementById("offer"); },
  get answer() { return document.getElementById("answer"); },
  get offerICE() { return document.getElementById("offerICE"); },
  get answerICE() { return document.getElementById("answerICE"); },
  get local() { return document.getElementById("local"); },
  get remote() { return document.getElementById("remote"); },
};


function rewrite(c, host, port) {
  c.candidate = c.candidate.replace(IP_PORT_REGEX, `${host} ${port}`);
  return c;
}

function run() {
  let offerer = new RTCPeerConnection();

  console.log("wtf?");

  navigator.mediaDevices.getUserMedia({video: true, audio: false})
    .then(stream => {
      console.log("got local stream");
      page.local.srcObject = stream;
      offerer.addStream(stream);
    });

  offerer.onicecandidate = e => {
    console.log("dropping local ICE candidate: " + JSON.stringify(e.candidate));
    return;
  }

  offerer.onnegotiationneeded = e => {
    offerer.createOffer().then(offer => {
      console.log("got offer");
      page.offer.value = offer.sdp;
      return offerer.setLocalDescription(offer);
    })
    .then(() => {
      console.log("setting static SDP answer");
      page.answer.value = ANSWER_SDP.sdp;
      return offerer.setRemoteDescription(ANSWER_SDP);
    })
    .then(() => {
      console.log("adding fake ICE candidates");
      let c = {"candidate": "candidate:0 1 UDP 2122121471 " + RELAY_IP + " " + RELAY_PORT + " typ host","sdpMid": "sdparta_0","sdpMLineIndex": 0};
      page.answerICE.value = JSON.stringify(c, null, 2) + "\n\n";
      return offerer.addIceCandidate(c);
    })
    .catch((error) => {
      console.log(error);
    })
  }
}

window.onload = () => {
  // Wire up actions
  page.run.onclick = run;

  // Clear the fields
  page.offer.value = "";
  page.answer.value = "";
  page.offerICE.value = "";
  page.answerICE.value = "";
};
