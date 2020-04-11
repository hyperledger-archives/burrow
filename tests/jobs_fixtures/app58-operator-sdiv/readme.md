* According to [Ethereum Yellow Paper](https://ethereum.github.io/yellowpaper/paper.pdf),
for signed division `a / b == sign(a/b) * floor(abs(a/b))`.
Therefore, `7 / 3`, `7 / (-3)`, `(-7) / 3` and `(-7) / (-3)` should give `2`, `-2`, `-2` and `2` respectively.