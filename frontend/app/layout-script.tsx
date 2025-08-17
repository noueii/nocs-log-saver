export function LayoutScript() {
  const script = `
    (function() {
      const theme = localStorage.getItem('theme') || 'dark';
      if (theme === 'dark') {
        document.documentElement.classList.add('dark');
      }
    })();
  `;
  
  return <script dangerouslySetInnerHTML={{ __html: script }} />;
}