function getArgs() {
    const args = [];
    var e = document.getElementById("format");
    var val = e.options[e.selectedIndex].value;
    args.push(val);
    e = document.getElementById("hash");
    val = e.options[e.selectedIndex].value;
    if (val != "none") {
        args.push(val);
    }
    val = document.querySelector('input[name="z"]:checked').value;
    if (val == "true") {
        args.push("z")
    }
    return args;
}

window.onload = () => {
    document.getElementById('butOpen').addEventListener('click', () => {

        window.showOpenFilePicker().then(handles => {
            for (const idx in handles) {
                const args = getArgs();
                args.unshift(handles[idx]);
                identify.apply(null, args).then(result => {
                    document.getElementById('results').value = result;
                }).catch((err) => {
                    document.getElementById('results').value = err;
                });
            };
        }
        );
    });
    document.getElementById('butDirectory').addEventListener('click', () => {
        window.showDirectoryPicker().then(handle => {
            const args = getArgs();
            args.unshift(handle);
            identify.apply(null, args).then(result => {
                document.getElementById('results').value = result;
            }).catch((err) => {
                document.getElementById('results').value = err;
            });
        }
        );
    });
}



