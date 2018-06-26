// https://blog.mozilla.org/webrtc/signaling-with-rtcsimpleconnection/

let gUMConfig = { "audio": false, "video": true };
const IP_PORT_REGEX = /\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\s+\d+/;
const RELAY_IP = "RELAY_IP_FROM_GO_SERVER";
const RELAY_PORT = "RELAY_PORT_FROM_GO_SERVER";
const IPV6_REGEX = RegExp('\:[0-9a-f]*\:[0-9a-fA-f]*','g');
const TCP_REGEX = RegExp('.*tcptype.*','g');

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
  let answerer = new RTCPeerConnection();
  
  console.log("wtf?");

  navigator.mediaDevices.getUserMedia({video: true, audio: false})
    .then(stream => {
      console.log("got local stream");
      page.local.srcObject = stream;
      offerer.addStream(stream);
    });
  
  answerer.ontrack = e => {
    console.log("got remote stream");
    page.remote.srcObject = e.streams[0];
  }
  
  offerer.onicecandidate = e => {
    if (!e.candidate ||
        IPV6_REGEX.test(e.candidate.candidate) ||
        TCP_REGEX.test(e.candidate.candidate)) {
      console.log("Ignoring offerer candidate: " + JSON.stringify(e.candidate));
      return;
    }
    console.log("got offerer ICE candidate: " + JSON.stringify(e.candidate));
    let candidate = rewrite(e.candidate, RELAY_IP, RELAY_PORT);
    page.offerICE.value += JSON.stringify(candidate, null, 2) + "\n\n";
    console.log("adding offerer ICE candidate: " + JSON.stringify(candidate));
    answerer.addIceCandidate(candidate);
  }

  answerer.onicecandidate = e => {
    if (!e.candidate ||
        IPV6_REGEX.test(e.candidate.candidate) ||
        TCP_REGEX.test(e.candidate.candidate)) {
      console.log("Ignoring answerer candidate: " + JSON.stringify(e.candidate));
      return;
    }
    console.log("got answerer ICE candidate: " + JSON.stringify(e.candidate));
    let candidate = rewrite(e.candidate, RELAY_IP, RELAY_PORT);
    page.answerICE.value += JSON.stringify(candidate, null, 2) + "\n\n";
    console.log("adding answerer ICE candidate: " + JSON.stringify(candidate));
    offerer.addIceCandidate(candidate);
  }

  offerer.onnegotiationneeded = e =>
    offerer.createOffer().then(offer => {
      console.log("got offer");
      page.offer.value = offer.sdp;
      offerer.setLocalDescription(offer);
      answerer.setRemoteDescription(offer);
    })
    .then(() => answerer.createAnswer()).then(answer => {
      console.log("got answer");
      page.answer.value = answer.sdp;
      answerer.setLocalDescription(answer);
      offerer.setRemoteDescription(answer);
    })
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
