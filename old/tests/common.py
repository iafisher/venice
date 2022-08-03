import textwrap


def S(s: str) -> str:
    """
    Dedents and strips a string so that it can be used in an indented multi-line string
    without extraneous whitespace, e.g.

      S(
          '''
          line 1
          line 2
          '''
      )

    is equal to

      "line 1\nline 2\n"
    """
    return textwrap.dedent(s).strip() + "\n"
