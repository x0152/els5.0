import { useEffect, useState } from 'react'
import { ArrowLeft, Loader2 } from 'lucide-react'
import { Avatar, Button, cn, Input, Mascot, Select, Textarea } from '@els/ui'
import { api } from '../lib/api'
import { useAuth } from '../auth/AuthContext'
import { WIZARD_TOUR, markTourDone } from './storage'

const LEVELS = [
  { code: 'A1', label: 'Beginner' },
  { code: 'A2', label: 'Elementary' },
  { code: 'B1', label: 'Intermediate' },
  { code: 'B2', label: 'Upper-intermediate' },
  { code: 'C1', label: 'Advanced' },
  { code: 'C2', label: 'Proficient' },
]

const LANGUAGES = [
  'Russian',
  'Ukrainian',
  'Belarusian',
  'Kazakh',
  'Spanish',
  'Portuguese',
  'French',
  'German',
  'Italian',
  'Polish',
  'Turkish',
  'Arabic',
  'Chinese',
  'Japanese',
  'Korean',
  'Hindi',
]

const STRICTNESS = [
  { value: 0.5, label: 'Easy', hint: 'forgiving' },
  { value: 1, label: 'Normal', hint: 'balanced' },
  { value: 2, label: 'Strict', hint: 'exact' },
] as const

const TOTAL_STEPS = 7

export function OnboardingWizard({ onDone }: { onDone: () => void }) {
  const { refresh } = useAuth()
  const [step, setStep] = useState(0)
  const [firstName, setFirstName] = useState('')
  const [lastName, setLastName] = useState('')
  const [englishLevel, setEnglishLevel] = useState('')
  const [nativeLanguage, setNativeLanguage] = useState('Russian')
  const [showTranslations, setShowTranslations] = useState(true)
  const [speechStrictness, setSpeechStrictness] = useState<0.5 | 1 | 2>(1)
  const [aboutMe, setAboutMe] = useState('')
  const [pictureUrl, setPictureUrl] = useState<string | undefined>(undefined)
  const [uploading, setUploading] = useState(false)
  const [saving, setSaving] = useState(false)
  const [err, setErr] = useState<string | null>(null)

  useEffect(() => {
    api.account
      .accountMe()
      .then((me) => {
        if (!me) return
        setFirstName(me.first_name)
        setLastName(me.last_name)
        if (me.english_level) setEnglishLevel(me.english_level)
        if (me.native_language) setNativeLanguage(me.native_language)
        setShowTranslations(me.show_translations ?? true)
        if (me.speech_strictness) setSpeechStrictness(me.speech_strictness as 0.5 | 1 | 2)
        setAboutMe(me.about_me ?? '')
        setPictureUrl(me.picture_url || undefined)
      })
      .catch(() => {})
  }, [])

  const canNext =
    (step !== 1 || (!!firstName.trim() && !!lastName.trim())) &&
    (step !== 3 || !!englishLevel)

  const uploadPicture = async (file: File) => {
    setErr(null)
    if (!/^image\//.test(file.type)) {
      setErr('An image must be uploaded')
      return
    }
    if (file.size > 5 * 1024 * 1024) {
      setErr('File is larger than 5 MB')
      return
    }
    setUploading(true)
    try {
      const form = new FormData()
      form.append('file', file)
      const res = await api.account.accountMeUploadPicture({ body: form as unknown as never })
      setPictureUrl(res?.picture_url || undefined)
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'Failed to upload photo')
    } finally {
      setUploading(false)
    }
  }

  const skip = () => {
    markTourDone(WIZARD_TOUR)
    onDone()
    void refresh()
  }

  const finish = async () => {
    setSaving(true)
    setErr(null)
    try {
      await api.account.accountUpdateProfile({
        body: {
          first_name: firstName,
          last_name: lastName,
          english_level: englishLevel,
          about_me: aboutMe,
          native_language: nativeLanguage,
          show_translations: showTranslations,
          speech_strictness: speechStrictness,
        },
      })
      markTourDone(WIZARD_TOUR)
      await refresh()
      onDone()
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'Failed to save profile')
      setSaving(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-end justify-center bg-black/40 backdrop-blur-sm sm:items-center sm:p-6">
      <div className="w-full max-w-xl overflow-hidden rounded-t-3xl bg-white shadow-2xl sm:rounded-3xl">
        <div className="h-1.5 bg-neutral-100">
          <div
            className="h-full bg-brand-500 transition-all duration-300"
            style={{ width: `${(step / (TOTAL_STEPS - 1)) * 100}%` }}
          />
        </div>

        <div className="max-h-[70dvh] overflow-y-auto p-6 sm:p-8">
          {step === 0 && (
            <div className="flex flex-col items-center py-4 text-center">
              <Mascot className="h-36 w-36" />
              <h2 className="mt-4 text-2xl font-bold text-neutral-900">Welcome to ELS!</h2>
              <p className="mt-2 max-w-sm text-sm text-neutral-600">
                Let&apos;s take a minute to set up your profile — the platform adapts lessons,
                translations and feedback to it.
              </p>
            </div>
          )}

          {step === 1 && (
            <div>
              <StepHeader
                title="What's your name?"
                subtitle="It appears on your profile and in the assistant's replies."
              />
              <div className="grid gap-4 sm:grid-cols-2">
                <Labeled label="First name">
                  <Input value={firstName} onChange={(e) => setFirstName(e.target.value)} autoFocus />
                </Labeled>
                <Labeled label="Last name">
                  <Input value={lastName} onChange={(e) => setLastName(e.target.value)} />
                </Labeled>
              </div>
            </div>
          )}

          {step === 2 && (
            <div>
              <StepHeader
                title="Add a photo"
                subtitle="Optional — makes your profile friendlier. You can change it any time."
              />
              <div className="flex flex-col items-center gap-3 py-4">
                <Avatar
                  src={pictureUrl}
                  name={`${firstName} ${lastName}`}
                  className="h-28 w-28 text-3xl"
                  onUpload={(file) => void uploadPicture(file)}
                  uploading={uploading}
                />
                <p className="text-xs text-neutral-400">PNG, JPEG or WebP, up to 5 MB</p>
                {err && <div className="text-xs text-rose-700">{err}</div>}
              </div>
            </div>
          )}

          {step === 3 && (
            <div>
              <StepHeader
                title="Your English level"
                subtitle="We'll adapt texts and exercises to it. Pick your best guess — you can change it any time."
              />
              <div className="grid grid-cols-2 gap-2 sm:grid-cols-3">
                {LEVELS.map((l) => {
                  const active = englishLevel.startsWith(l.code)
                  return (
                    <button
                      key={l.code}
                      type="button"
                      onClick={() => setEnglishLevel(`${l.code} (${l.label})`)}
                      className={cn(
                        'rounded-xl px-3 py-3 text-left ring-1 transition',
                        active
                          ? 'bg-brand-600 text-white ring-brand-600'
                          : 'bg-white text-neutral-800 ring-neutral-200 hover:bg-neutral-50',
                      )}
                    >
                      <div className="text-sm font-semibold">{l.code}</div>
                      <div className={cn('text-xs', active ? 'text-white/80' : 'text-neutral-500')}>
                        {l.label}
                      </div>
                    </button>
                  )
                })}
              </div>
            </div>
          )}

          {step === 4 && (
            <div>
              <StepHeader
                title="Your native language"
                subtitle="Used for translations and explanations across the platform."
              />
              <Labeled label="Native language">
                <Select
                  value={nativeLanguage}
                  onChange={setNativeLanguage}
                  options={(LANGUAGES.includes(nativeLanguage) || !nativeLanguage
                    ? LANGUAGES
                    : [nativeLanguage, ...LANGUAGES]
                  ).map((l) => ({ value: l, label: l }))}
                />
              </Labeled>
              <label className="mt-4 flex items-center gap-2">
                <input
                  type="checkbox"
                  checked={showTranslations}
                  onChange={(e) => setShowTranslations(e.target.checked)}
                  className="h-4 w-4 accent-brand-600"
                />
                <span className="text-sm text-neutral-700">
                  Show translations into my native language
                </span>
              </label>
            </div>
          )}

          {step === 5 && (
            <div>
              <StepHeader
                title="Pronunciation strictness"
                subtitle="How strictly Speaking exercises grade your pronunciation."
              />
              <div className="flex flex-wrap gap-2">
                {STRICTNESS.map((level) => (
                  <button
                    key={level.value}
                    type="button"
                    onClick={() => setSpeechStrictness(level.value)}
                    className={cn(
                      'rounded-xl px-4 py-3 text-left text-sm ring-1 transition',
                      speechStrictness === level.value
                        ? 'bg-brand-600 text-white ring-brand-600'
                        : 'bg-white text-neutral-700 ring-neutral-200 hover:bg-neutral-50',
                    )}
                  >
                    <span className="font-medium">{level.label}</span>
                    <span
                      className={cn(
                        'ml-1.5 text-xs',
                        speechStrictness === level.value ? 'text-white/80' : 'text-neutral-400',
                      )}
                    >
                      {level.hint}
                    </span>
                  </button>
                ))}
              </div>
            </div>
          )}

          {step === 6 && (
            <div>
              <StepHeader
                title="About you"
                subtitle="Optional — helps the AI pick topics and examples you'll actually enjoy."
              />
              <Textarea
                value={aboutMe}
                onChange={(e) => setAboutMe(e.target.value)}
                rows={5}
                placeholder="Your work, hobbies, why you're learning English…"
              />
              {err && <div className="mt-3 text-xs text-rose-700">{err}</div>}
            </div>
          )}
        </div>

        <div className="flex items-center justify-between border-t border-neutral-100 px-6 py-4 sm:px-8">
          <button
            type="button"
            onClick={skip}
            className="text-xs font-medium text-neutral-400 hover:text-neutral-600"
          >
            Skip for now
          </button>
          <div className="flex items-center gap-2">
            {step > 0 && (
              <Button variant="ghost" onClick={() => setStep(step - 1)} disabled={saving}>
                <ArrowLeft className="h-4 w-4" /> Back
              </Button>
            )}
            {step < TOTAL_STEPS - 1 ? (
              <Button variant="brand" onClick={() => setStep(step + 1)} disabled={!canNext}>
                {step === 0 ? "Let's go" : 'Next'}
              </Button>
            ) : (
              <Button variant="brand" onClick={finish} disabled={saving}>
                {saving && <Loader2 className="h-4 w-4 animate-spin" />} Finish
              </Button>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

function StepHeader({ title, subtitle }: { title: string; subtitle: string }) {
  return (
    <div className="mb-5">
      <h2 className="text-xl font-bold text-neutral-900">{title}</h2>
      <p className="mt-1 text-sm text-neutral-500">{subtitle}</p>
    </div>
  )
}

function Labeled({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <label className="block">
      <span className="mb-1 block text-[11px] font-semibold uppercase tracking-wider text-neutral-500">
        {label}
      </span>
      {children}
    </label>
  )
}
