import { useEffect, useState } from 'react'
import type { AISettingsView } from '../types'
import { api } from '../services/api'
import './AIIntegration.css'

export function AISettings() {
  const [settings, setSettings] = useState<AISettingsView | null>(null)
  const [key, setKey] = useState('')
  const [model, setModel] = useState('gemini-3.8-flash')
  const [loading, setLoading] = useState(true)
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState('')
  const [notice, setNotice] = useState('')
  const [confirmDelete, setConfirmDelete] = useState(false)
  async function load() {
    setLoading(true); setError('')
    try { const data = await api.getAISettings(); setSettings(data); setModel(data.model) }
    catch (err) { setError(err instanceof Error ? err.message : 'Pengaturan belum bisa dimuat.') }
    finally { setLoading(false) }
  }
  useEffect(() => {
    let cancelled = false
    api.getAISettings().then(data => { if (!cancelled) { setSettings(data); setModel(data.model) } })
      .catch(err => { if (!cancelled) setError(err instanceof Error ? err.message : 'Pengaturan belum bisa dimuat.') })
      .finally(() => { if (!cancelled) setLoading(false) })
    return () => { cancelled = true }
  }, [])
  async function save(event: React.FormEvent) {
    event.preventDefault(); setBusy(true); setError(''); setNotice('')
    try {
      const data = await api.saveAISettings({ api_key: key || undefined, model })
      setSettings(data); setKey(''); setNotice('Pengaturan tersimpan. Validitas key akan diperiksa ketika membuat draft.')
    } catch (err) { setError(err instanceof Error ? err.message : 'Pengaturan belum bisa disimpan.') }
    finally { setBusy(false) }
  }
  async function remove() {
    setBusy(true); setError(''); setNotice('')
    try { const data = await api.deleteAISettings(); setSettings(data); setModel(data.model); setKey(''); setConfirmDelete(false); setNotice('API key dihapus dari storage. Generator perlu key baru.') }
    catch (err) { setError(err instanceof Error ? err.message : 'Key belum bisa dihapus.') }
    finally { setBusy(false) }
  }
  return <section className="ai-section" aria-labelledby="ai-settings-title" aria-busy={loading || busy}>
    <h2 id="ai-settings-title">Pengaturan Gemini</h2>
    <p>API key disimpan terenkripsi di database. Hanya admin dapat mengganti atau menghapusnya; nilai tersimpan tidak ditampilkan kembali.</p>
    {loading ? <p role="status">Memuat pengaturan...</p> : !settings ? <button type="button" className="btn btn-secondary" onClick={load}>Muat ulang pengaturan</button> : <>
      <p role="status">Status key: <strong>{settings.configured ? 'Tersimpan' : 'Belum diatur'}</strong></p>
      {!settings.storage_ready && <p role="alert">Storage enkripsi belum siap. Konfigurasikan kunci enkripsi backend untuk menyimpan key.</p>}
      <form onSubmit={save}>
        <fieldset disabled={busy || !settings.storage_ready}>
          <legend className="visually-hidden">Kredensial dan model Gemini</legend>
          <div className="form-group"><label htmlFor="gemini-key">{settings.configured ? 'Ganti API key' : 'API key'}</label><input id="gemini-key" type="password" autoComplete="new-password" value={key} onChange={e => setKey(e.target.value)} required={!settings.configured} minLength={16} maxLength={512} aria-describedby="gemini-key-note" /><p id="gemini-key-note" className="ai-note">{settings.configured ? 'Kosongkan untuk mempertahankan key yang tersimpan.' : 'Masukkan key dari Google AI Studio.'}</p></div>
          <div className="form-group"><label htmlFor="gemini-model">Model Gemini</label><input id="gemini-model" value={model} onChange={e => setModel(e.target.value)} required pattern="gemini-[a-z0-9.\-]+" maxLength={100} /></div>
          <button type="submit" className="btn btn-primary">{busy ? 'Menyimpan...' : 'Simpan pengaturan AI'}</button>
        </fieldset>
      </form>
      {settings.configured && <div className="ai-key-actions">{confirmDelete ? <><p>Hapus key tersimpan? Pembuatan draft baru berhenti sampai key diganti.</p><button type="button" className="btn btn-secondary" disabled={busy} onClick={remove}>Konfirmasi hapus key</button><button type="button" className="btn btn-secondary" disabled={busy} onClick={() => setConfirmDelete(false)}>Batal</button></> : <button type="button" className="btn btn-secondary" disabled={busy} onClick={() => setConfirmDelete(true)}>Hapus API key</button>}</div>}
    </>}
    {error && <p className="alert alert-error" role="alert">{error}</p>}
    {notice && <p role="status">{notice}</p>}
  </section>
}
