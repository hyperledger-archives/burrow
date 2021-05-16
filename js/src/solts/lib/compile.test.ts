import assert from 'assert';
import { TokenizeLinks } from './compile';

describe('compilation helpers', function () {
  it('should tokenize links', () => {
    const links = {
      'dir/Errors.sol': {
        Errors: [],
      },
      'lib/Utils.sol': {
        Utils: [],
      },
    };

    const actual = TokenizeLinks(links);
    assert.equal(actual[0], 'dir/Errors.sol:Errors');
    assert.equal(actual[1], 'lib/Utils.sol:Utils');
  });
});
