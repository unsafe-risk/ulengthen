(function () {
    function main() {
        const lengthen_form = document.getElementById('lengthen-form');
        lengthen_form.onsubmit = function (e) {
            e.preventDefault();

            const input = document.getElementById('lengthen-form-url');
            const url = input.value;

            const xhr = new XMLHttpRequest();
            xhr.open('POST', '/new');
            xhr.setRequestHeader('Content-Type', 'application/json');
            xhr.onload = function (e) {
                if (xhr.status !== 200) {
                    alert(xhr.status + ': ' + xhr.statusText);
                    return;
                }
                const response = JSON.parse(xhr.responseText);
                document.getElementById('lengthen-result-box').hidden = false;
                const result = document.getElementById('lengthen-result');
                let url = window.location.href.split('?')[0].split('#')[0];
                if (url.endsWith('/')) {
                    url = url.substring(0, url.length - 1);
                }
                url = punycode.toUnicode(url); // convert punycode to unicode
                url += '/' + response.data;
                result.value = url;
            }

            xhr.send(JSON.stringify({ url: url }));
        };

        const copy_button = document.getElementById('copy-button');
        copy_button.onclick = function (e) {
            if (navigator.clipboard) {
                navigator.clipboard.writeText(document.getElementById('lengthen-result').value);
            } else {
                const result = document.getElementById('lengthen-result');
                result.select();
                document.execCommand('copy');
            }
        }

        const clear_button = document.getElementById('clear-button');
        clear_button.onclick = function (e) {
            document.getElementById('lengthen-result-box').hidden = true;
            document.getElementById('lengthen-form-url').value = '';
        }
    }

    main();
})();
