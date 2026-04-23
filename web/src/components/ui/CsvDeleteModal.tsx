import { useState, useRef, type DragEvent } from 'react'
import { Trash2, Download, FileText, AlertTriangle, AlertCircle, X } from 'lucide-react'
import { Modal } from './Modal'
import { Button } from './Button'
import type { DeleteResult } from './CsvImportModal'

interface CsvDeleteModalProps {
  open: boolean
  onClose: () => void
  title: string
  onDelete: (file: File) => Promise<DeleteResult>
}

function downloadDeleteTemplate(filename: string) {
  const rows = ['id', '00000000-0000-0000-0000-000000000000']
  const blob = new Blob([rows.join('\n')], { type: 'text/csv;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  a.click()
  URL.revokeObjectURL(url)
}

export function CsvDeleteModal({ open, onClose, title, onDelete }: CsvDeleteModalProps) {
  const [file, setFile] = useState<File | null>(null)
  const [dragging, setDragging] = useState(false)
  const [loading, setLoading] = useState(false)
  const [result, setResult] = useState<DeleteResult | null>(null)
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

  const handleDelete = async () => {
    if (!file) return
    setLoading(true)
    try {
      const res = await onDelete(file)
      setResult(res)
    } catch (e: unknown) {
      setResult({ total: 0, deleted: 0, errors: [{ row: 0, message: e instanceof Error ? e.message : 'Erro ao excluir' }] })
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
              onClick={() => downloadDeleteTemplate(`modelo-excluir-${title.toLowerCase().replace(/\s+/g, '-')}.csv`)}
            >
              Baixar Modelo
            </Button>
            <Button
              variant="danger"
              onClick={handleDelete}
              loading={loading}
              disabled={!file}
              icon={<Trash2 size={15} />}
            >
              Excluir
            </Button>
          </>
        )
      }
    >
      {result ? (
        <div className="space-y-4">
          <div className="flex items-center gap-3 p-4 bg-red-50 border border-red-200 rounded-xl">
            <Trash2 size={20} className="text-red-600 shrink-0" />
            <div>
              <p className="text-sm font-semibold text-red-800">{result.deleted} registro(s) excluído(s)</p>
              <p className="text-xs text-red-600">{result.total} linha(s) processada(s) no total</p>
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
          <div className="flex items-start gap-3 p-3 bg-red-50 border border-red-200 rounded-xl">
            <AlertTriangle size={16} className="text-red-600 shrink-0 mt-0.5" />
            <p className="text-sm text-red-700">
              <strong>Atenção:</strong> esta operação é irreversível. Os registros serão excluídos permanentemente.
              O CSV deve conter apenas a coluna <span className="font-mono text-xs bg-red-100 px-1 py-0.5 rounded">id</span> com os UUIDs a excluir.
            </p>
          </div>

          <div
            onDragOver={e => { e.preventDefault(); setDragging(true) }}
            onDragLeave={() => setDragging(false)}
            onDrop={handleDrop}
            onClick={() => inputRef.current?.click()}
            className={`
              relative border-2 border-dashed rounded-xl p-8 cursor-pointer transition-all text-center
              ${dragging ? 'border-red-400 bg-red-50' : 'border-gray-200 hover:border-red-300 hover:bg-gray-50'}
            `}
          >
            <input ref={inputRef} type="file" accept=".csv" className="hidden" onChange={handleFile} />
            {file ? (
              <div className="flex flex-col items-center gap-2">
                <FileText size={32} className="text-red-500" />
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
                <Trash2 size={32} className="text-gray-300" />
                <p className="text-sm text-gray-500">Arraste um arquivo CSV ou <span className="text-red-600 font-medium">clique para selecionar</span></p>
              </div>
            )}
          </div>

          <div>
            <p className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-1.5">Coluna obrigatória</p>
            <span className="px-2 py-0.5 bg-gray-100 text-gray-600 text-xs font-mono rounded">id</span>
          </div>
        </div>
      )}
    </Modal>
  )
}
