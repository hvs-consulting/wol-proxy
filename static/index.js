const elementButton = document.getElementById('bStart');
const elementParagraph = document.getElementById('pStart');

const spinnerHTML = '<span class="spinner-border spinner-border-sm" role="status" aria-hidden="true"></span> Starting... <span id="spanTimer"></span>s';

let timer = 0;
let elementSpanTimer = null;
let timerIntervalId = null;
let pollIntervalId = null;


function wakeAndStartSpinner(wakeTime, aliveRoute) {
    elementButton.setAttribute('disabled', true);

    fetch('/api/wake', { method: 'POST' }).catch((err) => console.error(err));

    timer = wakeTime;
    elementParagraph.innerHTML = spinnerHTML;
    elementSpanTimer = document.getElementById('spanTimer');
    elementSpanTimer.innerText = timer;

    // store interval references for cleanup
    timerIntervalId = setInterval(updateTimer, 1000);
    if (pollIntervalId === null) {
        pollIntervalId = setInterval(tryPoll, 5000, aliveRoute);
    }
}

function tryPoll(aliveRoute) {
    fetch(aliveRoute).then((res) => {
        if (res.status != 425) {
            window.location.reload(true);
        }
    });
}

function updateTimer() {
    timer -= 1;

    if (timer <= 0) {
        // cleanup after unsuccessful wakeup call
        // let poll live, if server needs just a little more time
        elementButton.removeAttribute('disabled');
        clearInterval(timerIntervalId);
        elementParagraph.innerHTML = '';
    }

    elementSpanTimer.innerText = timer;
}
