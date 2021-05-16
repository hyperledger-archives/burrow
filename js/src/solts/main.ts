#!/usr/bin/env node

import program from 'commander';
import fs from 'fs';
import { Readable } from 'stream';
import { Compiled, NewFile, Print } from './api';

const STDIN = '-';

async function ReadAll(file: Readable) {
  const chunks: Buffer[] = [];
  for await (const chunk of file) {
    chunks.push(chunk);
  }
  return Buffer.concat(chunks).toString();
}

type Combined = {
  contracts: Record<
    string,
    {
      abi?: string;
      bin?: string;
    }
  >;
  version: string;
};

async function main() {
  program
    .name('ts-sol')
    .arguments('<source> [destination]')
    .description('Generate typescript classes for an ABI. If no destination specified, print result to STDOUT.')
    .action(async function (src, dst) {
      let source: string;
      if (src === STDIN) {
        source = await ReadAll(Readable.from(process.stdin));
      } else {
        source = await ReadAll(Readable.from(fs.createReadStream(src)));
      }
      const input: Combined = JSON.parse(source);

      const compiled: Compiled[] = [];
      for (const k in input.contracts) {
        if (!input.contracts[k].abi) {
          throw new Error(`ABI not given for: ${k}`);
        }
        if (!input.contracts[k].bin) {
          throw new Error(`Bin not given for: ${k}`);
        }
        compiled.push({
          name: k,
          abi: JSON.parse(input.contracts[k].abi),
          bin: input.contracts[k].bin,
          links: [],
        });
      }

      const target = NewFile(compiled);
      dst ? fs.writeFileSync(dst, Print(...target)) : console.log(Print(...target));
    });

  try {
    await program.parseAsync(process.argv);
  } catch (err) {
    console.log(err);
    process.exit(1);
  }
}

main();
