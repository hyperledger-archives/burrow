import * as path from 'path';
import { build } from '../build';

// Build these before any tests tha may rely on the generated output
build(path.join(__dirname, '..', '..', '..', 'src', 'solts', 'sol'));
