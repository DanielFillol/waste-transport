import { ChevronLeft, ChevronRight } from 'lucide-react'

interface PaginationProps {
  page: number
  pages: number
  total: number
  limit: number
  onChange: (page: number) => void
}

export function Pagination({ page, pages, total, limit, onChange }: PaginationProps) {
  const from = (page - 1) * limit + 1
  const to = Math.min(page * limit, total)

  if (total === 0) return null

  return (
    <div className="flex items-center justify-between px-1 py-3">
      <p className="text-sm text-gray-500">
        Exibindo <span className="font-medium text-gray-700">{from}–{to}</span> de{' '}
        <span className="font-medium text-gray-700">{total}</span>
      </p>
      <div className="flex items-center gap-1">
        <button
          onClick={() => onChange(page - 1)}
          disabled={page <= 1}
          className="p-1.5 rounded-lg hover:bg-gray-100 disabled:opacity-30 disabled:cursor-not-allowed text-gray-600"
        >
          <ChevronLeft size={16} />
        </button>
        {Array.from({ length: Math.min(pages, 7) }, (_, i) => {
          const p = i + 1
          return (
            <button
              key={p}
              onClick={() => onChange(p)}
              className={`w-8 h-8 text-sm rounded-lg font-medium transition-colors
                ${p === page
                  ? 'bg-brand-500 text-white'
                  : 'hover:bg-gray-100 text-gray-600'}
              `}
            >
              {p}
            </button>
          )
        })}
        <button
          onClick={() => onChange(page + 1)}
          disabled={page >= pages}
          className="p-1.5 rounded-lg hover:bg-gray-100 disabled:opacity-30 disabled:cursor-not-allowed text-gray-600"
        >
          <ChevronRight size={16} />
        </button>
      </div>
    </div>
  )
}
