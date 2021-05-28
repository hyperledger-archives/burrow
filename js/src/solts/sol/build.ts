import * as path from 'path';
import { build } from '../build';

// Build these before any tests tha may rely on the generated output
const basePath = path.join(__dirname, '..', '..', '..', 'src', 'solts', 'sol');
build(basePath, { burrowImportPath: (file) => path.join(path.relative(file, basePath), '../index') }).catch((err) => {
  console.log(`Could not build solts test files: ${err}`, err);
  process.exit(1);
});
