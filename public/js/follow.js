const followButton = document.getElementById("follow-btn");

if (followButton) {
    followButton.addEventListener("click", function (event) {
        event.preventDefault();
        toggleFollow();
    });
}

function toggleFollow() {
    fetch(`/follow/${username}`, {
        method: 'POST',
        headers: {
            'X-Requested-With': 'XMLHttpRequest',
        }
    })
        .then(() => {
            followButton.innerText = followButton.innerText === "Follow" ? "Unfollow" : "Follow";
        })
        .catch(error => {
            console.error("Error toggling follow:", error);
        });
}