export type ManuscriptStatus =
  | 'draft'
  | 'submitted'
  | 'under_initial_review'
  | 'revision_needed'
  | 'under_re_review'
  | 'under_final_review'
  | 'rejected'
  | 'accepted'
  | 'scheduled'
  | 'published'
  | 'withdrawn'

export interface Manuscript {
  ID: string
  AuthorID: string
  SectionID: string
  Status: ManuscriptStatus
  ActiveVersion: number
  Version: number
  UpdatedAt: string
  Versions: Array<{
    Number: number
    Title: string
    Abstract: string
    Tags: string[]
    Locked: boolean
  }>
}

export interface ManuscriptPage {
  Items: Manuscript[]
  NextCursor: string
}

export interface PublicArticle {
  ID: string
  Title: string
  Abstract: string
  SectionID: string
  Edition: number
  Tags: string[]
  PublishedAt: string | null
}

export interface PublicationPage {
  Items: PublicArticle[]
  Total: number
}

export interface APIError {
  code: string
  message: string
  request_id: string
  fields?: Array<{ field: string; message: string }>
}
