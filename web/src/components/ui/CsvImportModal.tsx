import { useState, useRef, type DragEvent } from 'react'
import { Upload, Download, FileText, CheckCircle, AlertCircle, X } from 'lucide-react'
import { Modal } from './Modal'
import { Button } from './Button'

export interface ImportResult {
  total: number
  created: number
  updated?: number
  errors: { row: number; message: string }[]
}

export interface DeleteResult {
  total: number
  deleted: number
  errors: { row: number; message: string }[]
}

interface CsvImportModalProps {
  open: boolean
  onClose: () => void
  title: string
  templateHeaders: string[]
  templateExample: string[]
  onImport: (file: File) => Promise<ImportResult>
}

function downloadCsv(filename: string, headers: string[], example: string[]) {
  const rows = [headers.join(','), example.join(',')]
  const blob = new Blob([rows.join('\n')], { type: 'text/csv;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  a.click()
  URL.revokeObjectURL(url)
}

export function CsvImportModal({ open, onClose, title, templateHeaders, templateExample, onImport }: CsvImportModalProps) {
  const [file, setFile] = useState<File | null>(null)
  const [dragging, setDragging] = useState(false)
  const [loading, setLoading] = useState(false)
  const [result, setResult] = useState<ImportResult | null>(null)
  const inputRef = useRef<HTMLInputElement>(null)

  const reset = () => { setFile(null); setResult(null) }

  const handleClose = () => { reset(); onClose() }

  const handleDrop = (e: DragEvent) => {
    e.preventDefault()
    setDragging(false)
    const dropped = e.dataTransfer.files[0]
    if (dropped?.name.endsWith('.csv')) { setFile(dropped); setResult(null) }
  }

  const handleFile = (e: React.ChangeEvent<HTMLInputElement>) => {
    const f = e.target.files?.[0]
    if (f) { setFile(f); setResult(null) }
  }

  const handleImport = async () => {
    if (!file) return
    setLoading(true)
    try {
      const res = await onImport(file)
      setResult(res)
    } catch (e: unknown) {
      setResult({ total: 0, created: 0, errors: [{ row: 0, message: e instanceof Error ? e.message : 'Erro ao importar' }] })
    } finally {
      setLoading(false)
    }
  }

  return (
    <Modal
      open={open}
      onClose={handleClose}
      title={title}
      footer={
        result ? (
          <Button onClick={handleClose}>Fechar</Button>
        ) : (
          <>
            <Button
              variant="secondary"
              icon={<Download size={15} />}
              onClick={() => downloadCsv(`modelo-${title.toLowerCase().replace(/\s+/g, '-')}.csv`, templateHeaders, templateExample)}
            >
              Baixar Modelo
            </Button>
            <Button onClick={handleImport} loading={loading} disabled={!file} icon={<Upload size={15} />}>
              Importar
            </Button>
          </>
        )
      }
    >
      {result ? (
        <div className="space-y-4">
          <div className="flex items-center gap-3 p-4 bg-green-50 border border-green-200 rounded-xl">
            <CheckCircle size={20} className="text-green-600 shrink-0" />
            <div>
              <p className="text-sm font-semibold text-green-800">{result.created} registro(s) importado(s)</p>
              <p className="text-xs text-green-600">{result.total} linha(s) processada(s) no total</p>
            </div>
          </div>
          {result.errors.length > 0 && (
            <div>
              <p className="text-sm font-semibold text-gray-700 mb-2 flex items-center gap-1.5">
                <AlertCircle size={15} className="text-amber-500" />
                {result.errors.length} erro(s)
              </p>
              <div className="max-h-48 overflow-y-auto space-y-1.5">
                {result.errors.map((err, i) => (
                  <div key={i} className="flex gap-2 p-2 bg-amber-50 border border-amber-100 rounded-lg text-xs">
                    {err.row > 0 && <span className="text-amber-600 font-mono shrink-0">Linha {err.row}:</span>}
                    <span className="text-amber-800">{err.message}</span>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      ) : (
        <div className="space-y-4">
          <p className="text-sm text-gray-500">
            Faça o upload de um arquivo <span className="font-mono text-xs bg-gray-100 px-1 py-0.5 rounded">.csv</span> com os dados para importar em massa.
            Use o botão <strong>Baixar Modelo</strong> para obter o formato correto.
          </p>

          <div
            onDragOver={e => { e.preventDefault(); setDragging(true) }}
            onDragLeave={() => setDragging(false)}
            onDrop={handleDrop}
            onClick={() => inputRef.current?.click()}
            className={`
              relative border-2 border-dashed rounded-xl p-8 cursor-pointer transition-all text-center
              ${dragging ? 'border-brand-400 bg-brand-50' : 'border-gray-200 hover:border-brand-300 hover:bg-gray-50'}
            `}
          >
            <input ref={inputRef} type="file" accept=".csv" className="hidden" onChange={handleFile} />
            {file ? (
              <div className="flex flex-col items-center gap-2">
                <FileText size={32} className="text-brand-500" />
                <p className="text-sm font-medium text-gray-800">{file.name}</p>
                <p className="text-xs text-gray-400">{(file.size / 1024).toFixed(1)} KB</p>
                <button
                  type="button"
                  onClick={e => { e.stopPropagation(); reset() }}
                  className="mt-1 p-1 rounded-full hover:bg-gray-100 text-gray-400"
                >
                  <X size={14} />
                </button>
              </div>
            ) : (
              <div className="flex flex-col items-center gap-2">
                <Upload size={32} className="text-gray-300" />
                <p className="text-sm text-gray-500">Arraste um arquivo CSV ou <span className="text-brand-600 font-medium">clique para selecionar</span></p>
              </div>
            )}
          </div>

          <div>
            <p className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-1.5">Colunas do modelo</p>
            <div className="flex flex-wrap gap-1.5">
              {templateHeaders.map(h => (
                <span key={h} className="px-2 py-0.5 bg-gray-100 text-gray-600 text-xs font-mono rounded">{h}</span>
              ))}
            </div>
          </div>
        </div>
      )}
    </Modal>
  )
}
