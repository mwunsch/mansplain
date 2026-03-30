#!/bin/sh
set -eu

DOCS="$(cd "$(dirname "$0")" && pwd)"
MAN="$(dirname "$DOCS")/man"
VERSION="${VERSION:-$(git describe --tags --always 2>/dev/null || echo dev)}"

build_page() {
  src="$1"
  out="$2"
  name="$(basename "$src")"

  echo "  $name -> $out"

  # Inject version into .Os before rendering
  tmp=$(mktemp)
  sed "s/^\.Os mansplain$/.Os mansplain $VERSION/" "$src" > "$tmp"

  # Generate HTML fragment
  fragment=$(mandoc -Thtml -O 'fragment,man=https://man.openbsd.org/%N.%S' "$tmp")
  rm -f "$tmp"

  # Rewrite cross-reference links
  fragment=$(echo "$fragment" | sed \
    -e 's|href="https://man.openbsd.org/mansplain.1"|href="mansplain.1.html"|g' \
    -e 's|href="https://man.openbsd.org/mansplain.7"|href="index.html"|g' \
    -e 's|href="https://man.openbsd.org/ronn-format.7"|href="ronn-format.7.html"|g' \
    -e 's|href="https://man.openbsd.org/groff.1"|href="https://man7.org/linux/man-pages/man1/groff.1.html"|g' \
    -e 's|href="https://man.openbsd.org/manpath.1"|href="https://man7.org/linux/man-pages/man1/manpath.1.html"|g')

  title=$(echo "$fragment" | grep 'head-ltitle' | sed 's/.*>\(.*\)<.*/\1/' | head -1)
  [ -z "$title" ] && title="mansplain"

  cat > "$DOCS/$out" <<EOF
<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>$title</title>
  <link rel="stylesheet" href="style.css">
</head>
<body>
$fragment
<script>
document.querySelectorAll('div.Bd-indent').forEach(function(div) {
  var pre = div.querySelector('pre');
  if (!pre) return;
  var btn = document.createElement('button');
  btn.className = 'copy-btn';
  btn.textContent = 'copy';
  btn.addEventListener('click', function() {
    navigator.clipboard.writeText(pre.textContent.trim());
    btn.textContent = 'copied';
    setTimeout(function() { btn.textContent = 'copy'; }, 1500);
  });
  div.appendChild(btn);
});
</script>
</body>
</html>
EOF
}

echo "Building docs ($VERSION)..."
build_page "$MAN/mansplain.7" "index.html"
build_page "$MAN/mansplain.1" "mansplain.1.html"
build_page "$MAN/ronn-format.7" "ronn-format.7.html"
echo "Done."
