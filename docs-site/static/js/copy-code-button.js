const highlightClass = document.querySelectorAll('.highlight');

highlightClass.forEach(element => {
  const copyIcon = document.createElement('i');
  copyIcon.classList.add('fas', 'fa-copy', 'copy-icon');
  copyIcon.style.color = 'white';
  copyIcon.style.display = 'none';
  element.appendChild(copyIcon);

  element.addEventListener('mouseenter', () => {
    copyIcon.style.display = 'inline';
  });

  element.addEventListener('mouseleave', () => {
    copyIcon.style.display = 'none';
    copyIcon.classList.replace('fa-check', 'fa-copy');
  });

  copyIcon.addEventListener('click', async () => {
    const selection = window.getSelection();
    const range = document.createRange();
    const codeElements = element.querySelectorAll('code');
    const codeElement = codeElements.length > 1 ? codeElements[1] : codeElements[0];
    range.selectNodeContents(codeElement);
    selection.removeAllRanges();
    selection.addRange(range);
    const text = selection.toString();

    try {
      await navigator.clipboard.writeText(text);
      console.log('Text copied to clipboard');
      copyIcon.classList.replace('fa-copy', 'fa-check');
      selection.removeAllRanges();
    } catch (error) {
      console.error('Failed to copy: ', error);
    }
  });
});
