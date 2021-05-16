import * as fs from 'fs';
import * as path from 'path';
import * as solcv5 from 'solc_v5';
import * as solcv8 from 'solc_v8';
import { Compiled, newFile, printNodes, tokenizeLinks } from './api';
import { decodeOutput, encodeInput, importLocal, inputDescriptionFromFiles } from './lib/compile';

const solcCompilers = {
  v5: solcv5,
  v8: solcv8,
} as const;

/**
 * This is our Solidity -> Typescript code generation function, it:
 *  - Compiles Solidity source
 *  - Generates typescript code wrapping the Solidity contracts and functions that calls Burrow
 *  - Generates typescript code to deploy the contracts
 *  - Outputs the ABI files into bin to be later included in the distribution (for Vent and other ABI-consuming services)
 */
export function build(
  srcPathOrFiles: string | string[],
  binPath = 'bin',
  solcVersion: keyof typeof solcCompilers = 'v5',
): void {
  fs.mkdirSync(binPath, { recursive: true });
  const solidityFiles = getSourceFilesList(srcPathOrFiles);
  const inputDescription = inputDescriptionFromFiles(solidityFiles);
  const input = encodeInput(inputDescription);
  const solc = solcCompilers[solcVersion];
  const solcOutput = solc.compile(input, { import: importLocal });
  const output = decodeOutput(solcOutput);
  if (output.errors && output.errors.length > 0) {
    throw new Error(output.errors.map((err) => err.formattedMessage).join('\n'));
  }

  for (const filename of Object.keys(output.contracts)) {
    const compiled: Compiled[] = [];
    const solidity = output.contracts[filename];
    for (const contract of Object.keys(solidity)) {
      const comp = output.contracts[filename][contract];
      compiled.push({
        name: contract,
        abi: comp.abi,
        bin: comp.evm.bytecode.object,
        links: tokenizeLinks(comp.evm.bytecode.linkReferences),
      });
    }
    const target = filename.replace(/\.[^/.]+$/, '.abi.ts');
    // Write the ABIs emitted for each file to the name of that file without extension. We flatten into a single
    // directory because that's what burrow deploy has always done.

    for (const [name, contract] of Object.entries(solidity)) {
      fs.writeFileSync(path.join('bin', name + '.bin'), JSON.stringify(contract));
    }
    fs.writeFileSync(target, printNodes(...newFile(compiled)));
  }
}
function getSourceFilesList(srcPathOrFiles: string | string[]): string[] {
  if (typeof srcPathOrFiles === 'string') {
    return fs
      .readdirSync(srcPathOrFiles, { withFileTypes: true })
      .filter((f) => path.extname(f.name) === '.sol')
      .map((f) => path.join(srcPathOrFiles, f.name));
  }
  return srcPathOrFiles;
}
