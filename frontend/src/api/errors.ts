import type { ErrorCode } from './client'

export function getAPIErrorCode(value: unknown): ErrorCode | undefined {
  if (typeof value !== 'object' || value === null || !('error' in value)) return undefined
  const detail = value.error
  if (typeof detail !== 'object' || detail === null || !('code' in detail)) return undefined
  return typeof detail.code === 'string' ? (detail.code as ErrorCode) : undefined
}
