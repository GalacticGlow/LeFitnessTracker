document.addEventListener("DOMContentLoaded", () => {
    fetch("/allworkouts")
        .then(response => response.json())
        .then(res => {
            if (!res.success) {
                console.error("API error:", res.error);
                return;
            }
            const workouts = res.data; // this is your array

            const workoutTable = document.querySelector(".workouts-table").getElementsByTagName("tbody")[0];
            workoutTable.innerHTML = ""; // clear existing

            workouts.forEach(workout => {
                const row = workoutTable.insertRow();

                // Date
                const dateCell = row.insertCell(0);
                const date = new Date(workout.date);
                dateCell.textContent = date.toISOString().split("T")[0]; // → "2025-07-12"

                // Type
                const typeCell = row.insertCell(1);
                typeCell.textContent = workout.wtype;

                // View button
                const viewCell = row.insertCell(2);
                const viewBtn = document.createElement("button");
                viewBtn.textContent = "View";
                viewBtn.classList.add("view-btn");
                viewCell.appendChild(viewBtn);

                // Delete button
                const deleteCell = row.insertCell(3);
                const deleteBtn = document.createElement("button");
                deleteBtn.textContent = "Delete";
                deleteBtn.classList.add("delete-btn");
                deleteCell.appendChild(deleteBtn);

                deleteBtn.addEventListener("click", () => {
                    if (!confirm(`Delete workout on ${workout.date}?`)) return;

                    const dateOnly = workout.date.split("T")[0];

                    fetch(`/removeworkout/${encodeURIComponent(dateOnly)}`, {
                        method: "DELETE"
                    })
                        .then(response => response.json())
                        .then(res => {
                            if (!res.success) {
                                console.error("Delete failed:", res.error);
                                return;
                            }
                            // remove row from UI
                            row.remove();
                        })
                        .catch(err => console.error("Failed to delete workout:", err));
                });
            });
        })
        .catch(err => console.error("Failed to load workouts:", err));
});

document.addEventListener("DOMContentLoaded", () => {
    const modal = document.getElementById("workoutModal");
    const closeModalBtn = modal.querySelector(".close-modal-btn");
    const viewButtons = document.querySelectorAll(".view-btn");
    const addExerciseBtn = modal.querySelector(".add-exercise-btn");
    const deleteExerciseBtn = modal.querySelector(".delete-exercise-btn");
    const exerciseTable = modal.querySelector(".exercise-table tbody");
    const addWorkoutBtn = document.querySelector(".add-workout-btn");
    const mainSect = document.getElementById("main-section");
    const workoutTable = mainSect.querySelector(".workouts-table tbody")

    addWorkoutBtn.addEventListener("click", async() => {
        const date = prompt("Enter workout date (YYYY-MM-DD):");

        if (!date.match(/^\d{4}-(0[1-9]|1[0-2])-(0[1-9]|[12][0-9]|3[01])$/)) {
            alert("The date you entered is not valid. Please try again:");
            return;
        }

        const type = prompt("Enter workout type (Push, Pull, Legs, etc):");

        if (!date || !type) {
            alert("Please provide both date and type.");
            return;
        }

        const workout = {
            date: date,
            wtype: type,
            data: "{}" // empty JSON string for now
        };

        try {
            const response = await fetch("/addworkout", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(workout)
            });

            const result = await response.json();

            if (response.ok) {
                // Add a row to the table
                const row = document.createElement("tr");
                row.innerHTML = `
                <td contenteditable="true">${date}</td>
                <td contenteditable="true">${type}</td>
                <td>
                    <button class="view-btn">View</button>
                    <button class="delete-btn">Delete</button>
                </td>
            `;
                workoutTable.appendChild(row);

                console.log("Workout added:", result);
            } else {
                alert("Error: " + result.error);
            }
        } catch (err) {
            console.error("Failed to add workout:", err);
        }
    });

    viewButtons.forEach(btn => {
        btn.addEventListener("click", () => {
            modal.style.display = "flex";
        });
    });

    // Select all current delete buttons
    const deleteButtons = document.querySelectorAll(".delete-btn");

    // Open modal when clicking "View"
    workoutTable.addEventListener("click", (event) => {
        if (event.target.classList.contains("view-btn")) {
            modal.style.display = "flex";
        }
    });

    // Close modal
    closeModalBtn.addEventListener("click", () => {
        modal.style.display = "none";
    });

    // Add new exercise row
    addExerciseBtn.addEventListener("click", () => {
        const newRow = document.createElement("tr");
        newRow.innerHTML = `
            <td contenteditable="true">New Exercise</td>
            <td contenteditable="true">0</td>
            <td contenteditable="true">0</td>
            <td contenteditable="true">0kg</td>
            <td contenteditable="true">Notes...</td>
          `;
        exerciseTable.appendChild(newRow);
    });

    // Delete last exercise row
    deleteExerciseBtn.addEventListener("click", () => {
        if (exerciseTable.rows.length > 0) {
            exerciseTable.deleteRow(exerciseTable.rows.length - 1);
        }
    });

    // Close modal if clicking outside
    window.addEventListener("click", (event) => {
        if (event.target === modal) {
            modal.style.display = "none";
        }
    });
});