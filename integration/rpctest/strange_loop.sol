pragma solidity ^0.4.16;

contract StrangeLoop {
    int top = 23;
    int bottom = 34;
    int depth = 17;
    bool down = true;
    // indexed puts it in topic
    event ChangeLevel(
        bytes32 indexed direction,
        int indexed newDepth);

    function UpsieDownsie() public returns (int i) {
        i = depth;
        if (down) {
            if (depth < bottom) {
                depth++;
                emit ChangeLevel("Upsie!", depth);
                i = depth;
                this.UpsieDownsie();
            } else {
                down = false;
                i = depth;
                this.UpsieDownsie();
            }
        } else if (depth > top) {
            depth--;
            emit ChangeLevel("Downsie!", depth);
            i = depth;
            this.UpsieDownsie();
        } else {
            down = true;
            i = depth;
            return;
        }
    }
}