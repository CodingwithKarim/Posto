const username = document.body.dataset.username;
const pageInputForm = document.querySelector(".page-input-form");

if (pageInputForm) {
    pageInputForm.addEventListener("submit", function (event) {
        event.preventDefault();
        updatePageInput(event);
        
    });
}

function updatePageInput(event) {
    const form = event.target;
    const input = event.target.querySelector('.page-input');
    const pageNumber = parseInt(input.value.trim());
    const maxPage = parseInt(input.max.trim());

    if (pageNumber && pageNumber >= 1 && pageNumber <= maxPage) {
        const path = form.dataset.redirect;

        window.location.href = `${path}/?page=${pageNumber}`;
    }
}