import React, { useState } from "react";

const API_BASE = "http://localhost:4000/api";

export default function CreateDentistPage() {
  const [firstName, setFirstName] = useState("");
  const [lastName, setLastName] = useState("");
  const [sex, setSex] = useState("");
  const [image, setImage] = useState(null);
  const [preview, setPreview] = useState("");

  const [status, setStatus] = useState("Ready");
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleImage = (e) => {
    const file = e.target.files?.[0];

    if (!file) {
      setImage(null);
      setPreview("");
      return;
    }

    setImage(file);

    // ใช้สำหรับ preview เท่านั้น
    const previewUrl = URL.createObjectURL(file);
    setPreview(previewUrl);
  };

  const handleSubmit = async (e) => {
    e.preventDefault();

    if (!firstName || !lastName || !sex) {
      setStatus("Please complete all fields.");
      return;
    }

    if (!image) {
      setStatus("Please select a photo.");
      return;
    }

    const fullname = `${firstName} ${lastName}`;

    setIsSubmitting(true);
    setStatus("Registering dentist...");

    try {
      const formData = new FormData();

      formData.append("dentistName", fullname);
      formData.append("sex", sex);
      formData.append("image", image);

      const res = await fetch(`${API_BASE}/dentist`, {
        method: "POST",
        body: formData,
      });

      const data = await res.json();

      if (!res.ok) {
        throw new Error(data.error || "Registration failed");
      }

      console.log("Dentist ID:", data.dentistId);
      console.log("Presigned URL:", data.presignedUrl);

      setStatus(data.message);

      setFirstName("");
      setLastName("");
      setSex("");
      setImage(null);
      setPreview("");

    } catch (err) {
      setStatus(err.message || String(err));
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="create-page">
      <h2>Create Dentist</h2>

      <form className="dentist-form" onSubmit={handleSubmit}>

        <div className="row">
          <div className="field">
            <label>First Name</label>

            <input
              type="text"
              value={firstName}
              onChange={(e) => setFirstName(e.target.value)}
              required
            />
          </div>

          <div className="field">
            <label>Last Name</label>

            <input
              type="text"
              value={lastName}
              onChange={(e) => setLastName(e.target.value)}
              required
            />
          </div>
        </div>

        <div className="field">
          <label>Sex</label>

          <select
            value={sex}
            onChange={(e) => setSex(e.target.value)}
            required
          >
            <option value="">Select sex</option>
            <option value="male">Male</option>
            <option value="female">Female</option>
          </select>
        </div>

        <div className="field">
          <label>Photo</label>

          <input
            type="file"
            accept="image/jpeg,image/png,image/webp"
            onChange={handleImage}
          />
        </div>

        {preview && (
          <div className="preview">
            <img
              src={preview}
              alt="preview"
              style={{
                maxWidth: "300px",
                maxHeight: "300px",
              }}
            />
          </div>
        )}

        <div style={{ marginTop: 12 }}>
          <button type="submit" disabled={isSubmitting}>
            {isSubmitting ? "Submitting..." : "Submit"}
          </button>
        </div>

        <p style={{ marginTop: 12 }}>
          {status}
        </p>

      </form>
    </div>
  );
}