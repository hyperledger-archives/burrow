pragma solidity ^0.4.16;

contract StrangeLoop {
    int top = 23;
    int bottom = 34;
    int depth = 17;
    bool down = true;

    function UpsieDownsie() public returns (int i) {
        i = depth;
        if (down) {
            if (depth < bottom) {
                depth++;
                i = depth;
                this.UpsieDownsie();
            } else {
                down = false;
                i = depth;
                this.UpsieDownsie();
            }
        } else if (depth > top) {
            depth--;
            i = depth;
            this.UpsieDownsie();
        } else {
            down = true;
            i = depth;
            return;
        }
    }
}