const setupLikeToggle = () => {
  const likeButton = document.getElementById('likeButton');
  const likeCountSpan = document.getElementById('likeCount');

  if (!likeButton || !likeCountSpan) return;

  let currentCount = parseInt(likeCountSpan.getAttribute('data-count'), 10) || 0;

  likeButton.addEventListener('click', async (e) => {
    e.preventDefault();

    try {
      const postID = likeButton.getAttribute('data-post-id');
      const response = await fetch(`/blogpost/${postID}/like`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' }
      });

      const data = await response.json();

      if (!response.ok) {
        console.error(data.error || 'Error toggling like');
        return;
      }

      let likeIcon = likeButton.querySelector('i');

      if (!likeIcon) {
        console.warn("Like icon SVG not found, waiting...", likeIcon);
        return;
      }

      likeIcon.classList.toggle("fas", data.liked);
      likeIcon.classList.toggle("far", !data.liked);

      // **Update like count**
      currentCount = data.liked ? currentCount + 1 : currentCount - 1;
      likeCountSpan.setAttribute('data-count', currentCount);
      likeCountSpan.textContent = `${currentCount} ${currentCount === 1 ? "Like" : "Likes"}`;
    } catch (error) {
      console.error('Error:', error);
    }
  });
};

setupLikeToggle();
