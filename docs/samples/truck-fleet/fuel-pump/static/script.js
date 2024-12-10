function updatePumpState(pumpId) {
    fetch(`http://localhost:8000/pump_truck?pump=${pumpId}`)
        .then(response => response.json())
        .then(data => {
            const pump = document.getElementById(`pump-${pumpId}`);
            const fill = pump.querySelector('.fill');
            const percentage = pump.querySelector('.percentage');
            const imagePlaceholder = pump.querySelector('.image-placeholder');
            const pumpName = pump.querySelector('.pump-name');

            if (data === null) {
                fill.style.height = '0%';
                percentage.textContent = '0%';
                imagePlaceholder.textContent = 'Image';
                pumpName.textContent = `Pump ${pumpId}`;
            } else {
                const fillPercentage = data.current_fuel;
                fill.style.height = fillPercentage + '%';
                percentage.textContent = fillPercentage + '%';
                imagePlaceholder.innerHTML = `<img src="${data.image}" alt="Pump Image" style="width: 100%; height: 100%; border-radius: 10px;">`;
                pumpName.textContent = data.name;
            }
        });
}

document.querySelectorAll('.fuel-up').forEach(button => {
    button.addEventListener('click', function() {
        const pumpId = this.getAttribute('data-id');
        fetch(`http://localhost:8000/start_fueling?pump=${pumpId}`, {
            method: 'POST',
            headers: {
            'Content-Type': 'application/json'
            }
        });
    });
});

// Initial polling to update the state of all pumps
setInterval(() => {
    for (let i = 1; i <= 3; i++) {
        updatePumpState(i);
    }
}, 1000); // Poll every second