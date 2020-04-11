* According to [Ethereum Yellow Paper](https://ethereum.github.io/yellowpaper/paper.pdf),
$a/b=sign(a/b)\lfloor |a/b|\rfloor$.
Therefore, `7 / 3`, `7 / (-3)`, `(-7) / 3` and `(-7) / (-3)` should give `2`, `-2`, `-2` and `2` respectively.