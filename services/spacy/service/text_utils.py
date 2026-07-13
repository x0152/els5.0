import re

HTML_TAG_RE = re.compile(r"<[^>]+>")


def strip_html(content: str) -> str:
    result = HTML_TAG_RE.sub(lambda m: " " * len(m.group()), content)
    return re.sub(r"\[\d{1,2}:\d{2}\.\d{2}\]", lambda m: " " * len(m.group()), result)


def strip_html_tags(content: str) -> str:
    return HTML_TAG_RE.sub("", content)


def strip_html_clean(content: str) -> str:
    text = content
    if ">" in text and "<" not in text.split(">")[0]:
        text = text.split(">", 1)[-1]
    if "<" in text and ">" not in text.rsplit("<", 1)[-1]:
        text = text.rsplit("<", 1)[0]
    text = HTML_TAG_RE.sub(" ", text)
    return " ".join(text.split())
