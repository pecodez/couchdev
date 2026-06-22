import { describe, it, expect } from 'vitest'
import { extractRepoName, isValidProjectName } from './utils.js'

describe('extractRepoName', () => {
  it('returns empty string for empty input', () => {
    expect(extractRepoName('')).toBe('')
  })

  it('extracts name from SSH URL with .git suffix', () => {
    expect(extractRepoName('git@github.com:org/my-project.git')).toBe('my-project')
  })

  it('extracts name from HTTPS URL with .git suffix', () => {
    expect(extractRepoName('https://github.com/org/repo-name.git')).toBe('repo-name')
  })

  it('extracts name from HTTPS URL without .git suffix', () => {
    expect(extractRepoName('https://github.com/org/repo')).toBe('repo')
  })

  it('strips trailing slash before extracting name', () => {
    expect(extractRepoName('https://github.com/org/repo/')).toBe('repo')
  })

  it('handles nested group paths in SSH URL', () => {
    expect(extractRepoName('git@gitlab.com:group/subgroup/repo.git')).toBe('repo')
  })
})

describe('isValidProjectName', () => {
  it('accepts hyphenated names', () => {
    expect(isValidProjectName('my-project')).toBe(true)
  })

  it('accepts underscore names', () => {
    expect(isValidProjectName('my_project')).toBe(true)
  })

  it('accepts mixed-case alphanumeric', () => {
    expect(isValidProjectName('MyProject123')).toBe(true)
  })

  it('accepts single character', () => {
    expect(isValidProjectName('a')).toBe(true)
  })

  it('rejects names with spaces', () => {
    expect(isValidProjectName('my project')).toBe(false)
  })

  it('rejects names starting with a dash', () => {
    expect(isValidProjectName('-project')).toBe(false)
  })

  it('rejects names with dots', () => {
    expect(isValidProjectName('my.project')).toBe(false)
  })

  it('rejects empty string', () => {
    expect(isValidProjectName('')).toBe(false)
  })
})
