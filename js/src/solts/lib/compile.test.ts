import assert from 'assert';
import { tokenizeLinks } from './compile';

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

    const actual = tokenizeLinks(links);
    assert.equal(actual[0], 'dir/Errors.sol:Errors');
    assert.equal(actual[1], 'lib/Utils.sol:Utils');
  });
});
