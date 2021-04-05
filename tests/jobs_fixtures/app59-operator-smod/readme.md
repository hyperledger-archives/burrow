* According to [Ethereum Yellow Paper](https://ethereum.github.io/yellowpaper/paper.pdf),
for signed modulo `a % b == sign(a) * (abs(a) % abs(b))`.
Therefore, `7 % 3`, `7 % (-3)`, `(-7) % 3` and `(-7) % (-3)` should give `1`, `1`, `-1` and `-1` respectively.