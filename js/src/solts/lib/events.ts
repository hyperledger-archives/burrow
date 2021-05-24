import * as ts from 'typescript';
import { factory } from 'typescript';
import { decodeName } from './decoder';
import { EventErrParameter, eventName } from './provider';
import { ContractMethodsList, getRealType, InputOutput, sha3 } from './solidity';
import {
  arrowFuncT,
  constObject,
  createCall,
  createParameter,
  declareConstant,
  EqualsGreaterThanToken,
  EventStreamType,
  EventType,
  ExportToken,
  listenerForName,
  NumberType,
  prop,
  ReturnType,
  SignalType,
  TType,
  UnknownType,
  VoidType,
} from './syntax';

export const eventsName = factory.createIdentifier('events');
export const eventNameTypeName = factory.createIdentifier('EventName');
export const BoundsType = factory.createUnionTypeNode([
  ...['first', 'latest', 'stream'].map((s) => factory.createLiteralTypeNode(factory.createStringLiteral(s))),
  NumberType,
]);
export const CallbackReturnType = factory.createUnionTypeNode([SignalType, VoidType]);

const typedListenerName = factory.createIdentifier('TypedListener');

const getLogName = factory.createIdentifier('getLog');
const getDataName = factory.createIdentifier('getData_asU8');
const getTopicsName = factory.createIdentifier('getTopicsList_asU8');
const taggedPayloadName = factory.createIdentifier('TaggedPayload');
const solidityEventName = factory.createIdentifier('SolidityEvent');
const eventRegistryName = factory.createIdentifier('EventRegistry');
const signatureName = factory.createIdentifier('signature');
const taggedName = factory.createIdentifier('tagged');
const payloadName = factory.createIdentifier('payload');

export function eventSignature(name: string, inputs: InputOutput[]): string {
  return `${name}(${inputs.map(({ type }) => type).join(',')})`;
}

export function eventSigHash(name: string, inputs: InputOutput[]): string {
  return sha3(eventSignature(name, inputs));
}

function callGetLogFromEvent(event: ts.Expression): ts.CallExpression {
  return createCall(prop(event, getLogName, true));
}

export function callGetDataFromEvent(event: ts.Expression): ts.CallExpression {
  return createCall(prop(callGetLogFromEvent(event), getDataName, true));
}

export function callGetTopicsFromEvent(event: ts.Expression): ts.CallExpression {
  return createCall(prop(callGetLogFromEvent(event), getTopicsName, true));
}

export function createListenerForFunction(clientName: ts.Identifier, addressName: ts.Identifier): ts.ArrowFunction {
  const eventNamesName = factory.createIdentifier('eventNames');
  const typedListener = factory.createTypeReferenceNode(typedListenerName, [TType]);
  return arrowFuncT(
    [createParameter(eventNamesName, factory.createArrayTypeNode(TType))],
    factory.createTypeReferenceNode(eventNameTypeName),
    typedListener,
    factory.createAsExpression(
      // The intermediate as unknown expression is not needed if tsconfig is in 'strict mode' but otherwise tsc complains
      // that the types do not sufficiently overlap
      factory.createAsExpression(
        createCall(listenerForName, [clientName, addressName, eventsName, decodeName, eventNamesName]),
        UnknownType,
      ),
      typedListener,
    ),
  );
}

export function createListener(clientName: ts.Identifier, addressName: ts.Identifier): ts.AsExpression {
  const eventNamesType = factory.createTypeReferenceNode(eventNameTypeName);
  const typedListener = factory.createTypeReferenceNode(typedListenerName, [eventNamesType]);
  return factory.createAsExpression(
    createCall(listenerForName, [
      clientName,
      addressName,
      eventsName,
      decodeName,
      factory.createAsExpression(
        createCall(prop('Object', 'keys'), [eventsName]),
        factory.createArrayTypeNode(eventNamesType),
      ),
    ]),
    typedListener,
  );
}

export function declareEvents(events: ContractMethodsList): ts.VariableStatement {
  return declareConstant(
    eventsName,
    constObject(
      events.map((a) => {
        const signature = a.signatures[0];
        return factory.createPropertyAssignment(
          a.name,
          constObject([
            factory.createPropertyAssignment(signatureName, factory.createStringLiteral(signature.hash)),
            factory.createPropertyAssignment(
              taggedName,
              factory.createArrowFunction(
                undefined,
                undefined,
                signature.inputs.map((i) => createParameter(i.name, getRealType(i.type))),
                undefined,
                EqualsGreaterThanToken,
                constObject([
                  factory.createPropertyAssignment('name', factory.createStringLiteral(a.name)),
                  factory.createPropertyAssignment(
                    payloadName,
                    constObject(
                      signature.inputs.map(({ name }) =>
                        factory.createPropertyAssignment(name, factory.createIdentifier(name)),
                      ),
                    ),
                  ),
                ]),
              ),
            ),
          ]),
        );
      }),
    ),
  );
}

export function eventTypes(): ts.TypeAliasDeclaration[] {
  const tExtendsEventName = factory.createTypeParameterDeclaration(
    factory.createIdentifier('T'),
    factory.createTypeReferenceNode(eventNameTypeName, undefined),
    undefined,
  );
  const tType = factory.createTypeReferenceNode(factory.createIdentifier('T'), undefined);
  return [
    factory.createTypeAliasDeclaration(
      undefined,
      undefined,
      eventRegistryName,
      undefined,
      factory.createTypeQueryNode(eventsName),
    ),
    factory.createTypeAliasDeclaration(
      undefined,
      [factory.createModifier(ts.SyntaxKind.ExportKeyword)],
      eventNameTypeName,
      undefined,
      factory.createTypeOperatorNode(
        ts.SyntaxKind.KeyOfKeyword,
        factory.createTypeReferenceNode(eventRegistryName, undefined),
      ),
    ),
    factory.createTypeAliasDeclaration(
      undefined,
      [ExportToken],
      taggedPayloadName,
      [tExtendsEventName],
      factory.createIntersectionTypeNode([
        factory.createTypeReferenceNode(ReturnType, [
          factory.createIndexedAccessTypeNode(
            factory.createIndexedAccessTypeNode(factory.createTypeReferenceNode(eventRegistryName, undefined), tType),
            factory.createLiteralTypeNode(factory.createStringLiteral(taggedName.text)),
          ),
        ]),
        factory.createTypeLiteralNode([factory.createPropertySignature(undefined, eventName, undefined, EventType)]),
      ]),
    ),
    factory.createTypeAliasDeclaration(
      undefined,
      [ExportToken],
      solidityEventName,
      [tExtendsEventName],
      factory.createIndexedAccessTypeNode(
        factory.createTypeReferenceNode(taggedPayloadName, [tType]),
        factory.createLiteralTypeNode(factory.createStringLiteral(payloadName.text)),
      ),
    ),
    factory.createTypeAliasDeclaration(
      undefined,
      [ExportToken],
      typedListenerName,
      [factory.createTypeParameterDeclaration('T', factory.createTypeReferenceNode(eventNameTypeName))],
      factory.createFunctionTypeNode(
        [],
        [
          createParameter(
            'callback',
            factory.createFunctionTypeNode(
              [],
              [
                EventErrParameter,
                createParameter(
                  eventName,
                  factory.createTypeReferenceNode(taggedPayloadName, [TType]),
                  undefined,
                  true,
                ),
              ],
              VoidType,
            ),
          ),
          createParameter('start', BoundsType, undefined, true),
          createParameter('end', BoundsType, undefined, true),
        ],
        EventStreamType,
      ),
    ),
  ];
}
